package service

import (
	"be-assesment-product/lib/utils"
	"be-assesment-product/model"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"be-assesment-product/db"
	"be-assesment-product/redis"

	redis2 "github.com/go-redis/redis/v8"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductService interface {
	AddProduct(ctx context.Context, p *model.CreateProductRequest) error
	ListProducts(ctx context.Context, req *model.ListProductRequest) (*model.ListProductResponse, error)
}

type productService struct {
	dbProvider *db.GormProvider
	logger     *logrus.Logger
	redis      *redis.Redis
}

func NewProductService(dbProvider *db.GormProvider, logger *logrus.Logger, redis *redis.Redis) ProductService {
	return &productService{
		dbProvider: dbProvider,
		logger:     logger,
		redis:      redis,
	}
}

func setPagination(page int32, limit int32) *model.PaginationResponse {
	res := &model.PaginationResponse{
		Limit: 10,
		Page:  1,
	}

	if limit == 0 && page == 0 {
		res.Limit = -1
		res.Page = -1
		return res
	} else {
		res.Limit = limit
		res.Page = page
	}

	if res.Page == 0 {
		res.Page = 1
	}

	switch {
	case res.Limit > 100:
		res.Limit = 100
	case res.Limit <= 0:
		res.Limit = 10
	}

	return res
}

func (s *productService) AddProduct(ctx context.Context, req *model.CreateProductRequest) error {
	remarkLength, _ := strconv.Atoi(utils.GetEnv("REMARK_LENGTH", "500"))
	nameLength, _ := strconv.Atoi(utils.GetEnv("PRODUCT_NAME_LENGTH", "255"))

	s.logger.Info("Start Validation for req ", req)

	if utils.IsEmptyString(req.Name) {
		return status.Errorf(codes.Aborted, "product name is empty")
	}
	if len(req.Name) > nameLength {
		return status.Errorf(codes.Aborted, "product name maximum characters is %d", nameLength)
	}

	if !utils.IsValidProductName(req.Name) {
		return status.Errorf(codes.Aborted, "characters not allowed in column product name")
	}

	s.logger.Info("Start GetProductByName ", req.Name)

	exist, err := s.dbProvider.GetProductByName(ctx, req.Name)
	if err != nil {
		s.logger.Error("err GetProductByName ", err)
		return status.Error(codes.Internal, "Internal Server Error")
	}

	s.logger.Info("Success GetProductByName ", req.Name)

	if exist.Name != "" {
		return status.Errorf(codes.Aborted, "product name already exist")
	}

	if req.Price <= 0 {
		return status.Errorf(codes.Aborted, "minimum price is 1")
	}

	if req.Quantity <= 0 {
		return status.Errorf(codes.Aborted, "minimum quantity is 1")
	}

	if strings.TrimSpace(req.Description) == "" {
		req.Description = "-"
	} else {
		if len(req.Description) > remarkLength {
			return status.Errorf(codes.Aborted, fmt.Sprintf("%s maximum characters is %d", req.Description, remarkLength))
		}
		if !utils.IsValidCharacter(req.Description) {
			return status.Errorf(codes.Aborted, fmt.Sprintf("characters not allowed in field Description", req.Description))
		}
	}

	s.logger.Info("Start InsertProduct with data ", req)

	err = s.dbProvider.InsertProduct(ctx, req)
	if err != nil {
		s.logger.Error("err InsertProduct ", err)
		return status.Error(codes.Internal, "Internal Server Error")
	}

	s.logger.Info("Success InsertProduct")

	return nil
}

func (s *productService) ListProducts(ctx context.Context, req *model.ListProductRequest) (*model.ListProductResponse, error) {

	s.logger.Info("Start ListProducts with req : ", req)
	s.logger.Info("Start Decode Filter")

	decodeQuery, err := base64.RawStdEncoding.DecodeString(req.Query)
	if err != nil {
		s.logger.Error("err DecodeString ", err)
		return nil, nil
	}

	s.logger.Info("Success Decode Query")

	pagination := setPagination(req.Page, req.Limit)

	allowedColumns := map[string]bool{
		"created_at": true,
		"price":      true,
		"name":       true,
		"product_id": true,
		"quantity":   true,
	}
	allowedDirections := map[string]bool{
		"asc":  true,
		"desc": true,
	}

	column := strings.ToLower(req.Sort)
	direction := strings.ToLower(req.Dir)

	if !allowedColumns[column] {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Argument")
	}
	if !allowedDirections[direction] {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Argument")
	}

	sort := &model.Sort{
		Column:    column,
		Direction: direction,
	}

	sqlBuilder := &db.QueryBuilder{
		CollectiveAnd: string(decodeQuery),
		Sort:          sort,
	}

	redisKey := fmt.Sprintf("PRODUCT-REQ:sort=%s&dir=%s&page=%d&limit=%d&query=%s",
		req.Sort, req.Dir, req.Page, req.Limit, req.Query)

	s.logger.Info("Start StoreDataWithTimeout with key", redisKey)

	product, err := s.StoreDataWithTimeout(ctx, redisKey, pagination, sqlBuilder, sort)
	if err != nil {
		return nil, err
	}

	s.logger.Info("Success StoreDataWithTimeout ", product)
	s.logger.Info("Start making response")

	products := &model.ListProductResponse{
		Error:      false,
		Code:       http.StatusOK,
		Message:    "Success",
		Data:       product,
		Pagination: pagination,
	}

	return products, nil

}

func (s *productService) StoreDataWithTimeout(ctx context.Context, key string, pagination *model.PaginationResponse, sql *db.QueryBuilder, sort *model.Sort) (data []*model.Product, err error) {
	rdb := s.redis.GetClient()

	s.logger.Info("start get or store list product wwith key : ", key)

	cachedData, err := rdb.Get(ctx, key).Bytes()
	if err == nil {
		var product []*model.Product
		err = json.Unmarshal(cachedData, &product)
		if err != nil {
			s.logger.Error("err failed to unmarshal data from Redis:", err)
			return nil, status.Error(codes.Internal, "Internal Error")
		}

		if len(product) > 0 {
			return product, nil
		}

	} else if !errors.Is(err, redis2.Nil) {
		s.logger.Error("err : ", err)
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	s.logger.Info("Start ListProduct")
	product, err := s.dbProvider.ListProduct(ctx, pagination, sql, sort)
	if err != nil {
		s.logger.Error("err ListProduct ", err)
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	dataBytes, err := json.Marshal(product)
	if err != nil {
		s.logger.Error("Failed to marshal data", err)
		return nil, status.Error(codes.Internal, "Error serializing data to cache")
	}

	err = rdb.Set(ctx, key, dataBytes, 1*time.Minute).Err()
	if err != nil {
		s.logger.Error("Failed to set data in Redis", err)
		return nil, status.Error(codes.Internal, "Error storing data to cache")
	}

	return product, nil
}
