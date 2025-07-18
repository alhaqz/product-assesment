package db

import (
	"context"
	"errors"
	"time"

	"be-assesment-product/model"

	"gorm.io/gorm"
)

func (p *GormProvider) InsertProduct(ctx context.Context, req *model.CreateProductRequest) error {
	timeoutctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	query := p.db_main.WithContext(timeoutctx).Debug().Table("public.products")

	now := time.Now()
	data := &model.Product{
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		Quantity:    req.Quantity,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	if err := query.Create(data).Error; err != nil {
		return err
	}

	return nil
}

func (p *GormProvider) GetProductByName(ctx context.Context, productName string) (*model.Product, error) {
	timeoutctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	var data model.Product

	err := p.db_main.WithContext(timeoutctx).
		Table("public.products").
		Where("name = ?", productName).
		First(&data).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.Product{}, nil
		}

		return nil, err
	}

	return &data, nil
}

func (p *GormProvider) ListProduct(ctx context.Context, pagination *model.PaginationResponse, sql *QueryBuilder, sort *model.Sort) (data []*model.Product, err error) {
	timeoutctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	query := p.db_main.WithContext(timeoutctx).Debug().Table("public.products")

	query = query.Scopes(
		QueryScoop(sql.CollectiveAnd),
	)

	query = query.Scopes(Paginate(data, pagination, query))
	query = query.Scopes(
		Sort(sort),
	)

	if err := query.Debug().Find(&data).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	return data, nil
}
