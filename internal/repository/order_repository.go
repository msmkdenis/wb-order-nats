package repository

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/internal/model"
	"github.com/msmkdenis/wb-order-nats/internal/storage/db"
	"github.com/msmkdenis/wb-order-nats/pkg/apperr"
)

//go:embed queries/insert_delivery.sql
var insertDelivery string

//go:embed queries/insert_item.sql
var insertItem string

//go:embed queries/insert_order.sql
var insertOrder string

//go:embed queries/insert_payment.sql
var insertPayment string

//go:embed queries/select_full_order_by_id.sql
var selectFullOrder string

//go:embed queries/select_all_full_orders.sql
var selectAllFullOrders string

type OrderRepository struct {
	postgresPool *db.PostgresPool
	logger       *zap.Logger
}

func NewOrderRepository(postgresPool *db.PostgresPool, logger *zap.Logger) *OrderRepository {
	return &OrderRepository{
		postgresPool: postgresPool,
		logger:       logger,
	}
}

func (r *OrderRepository) Insert(ctx context.Context, o model.Order) error {
	d := o.Delivery
	p := o.Payment

	tx, err := r.postgresPool.DB.Begin(context.Background())
	if err != nil {
		r.logger.Info("Error while staring transaction", zap.String("error", err.Error()))
		return err
	}
	defer tx.Rollback(context.Background())

	order, err := tx.Prepare(context.Background(), "insertOrder", insertOrder)
	if err != nil {
		return apperr.NewValueError("unable to prepare query", apperr.Caller(), err)
	}

	delivery, err := tx.Prepare(context.Background(), "insertDelivery", insertDelivery)
	if err != nil {
		return apperr.NewValueError("unable to prepare query", apperr.Caller(), err)
	}

	payment, err := tx.Prepare(context.Background(), "insertPayment", insertPayment)
	if err != nil {
		return apperr.NewValueError("unable to prepare query", apperr.Caller(), err)
	}

	item, err := tx.Prepare(context.Background(), "insertItem", insertItem)
	if err != nil {
		return apperr.NewValueError("unable to prepare query", apperr.Caller(), err)
	}

	batch := &pgx.Batch{}
	batch.Queue(order.Name, o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature, o.CustomerID,
		o.DeliveryService, o.Shardkey, o.SmID, o.DateCreated, o.OofShard)

	batch.Queue(delivery.Name, d.Name, d.Phone, d.Zip, d.City, d.Address, d.Region, d.Email, o.OrderUID)

	batch.Queue(payment.Name, p.Transaction, p.RequestID, p.Currency, p.Provider, p.Amount, p.PaymentDt, p.Bank,
		p.DeliveryCost, p.GoodsTotal, p.CustomFee, o.OrderUID)

	for _, i := range o.Items {
		batch.Queue(item.Name, i.ChrtID, i.TrackNumber, i.Price, i.Rid, i.Name, i.Sale, i.Size, i.TotalPrice,
			i.NmID, i.Brand, i.Status, o.OrderUID)
	}

	result := tx.SendBatch(context.Background(), batch)

	err = result.Close()
	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (r *OrderRepository) SelectById(ctx context.Context, orderId string) (*model.Order, error) {
	rows, err := r.postgresPool.DB.Query(ctx, selectFullOrder, orderId)
	if err != nil {
		r.logger.Info("error", zap.Error(err))
		return nil, err
	}

	order, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Order])
	if err != nil {
		r.logger.Info("error", zap.Error(err))
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) SelectAll(ctx context.Context) ([]model.Order, error) {
	rows, err := r.postgresPool.DB.Query(ctx, selectAllFullOrders)
	if err != nil {
		r.logger.Info("error", zap.Error(err))
		return nil, err
	}

	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Order])
	if err != nil {
		r.logger.Info("error", zap.Error(err))
		return nil, err
	}

	return orders, nil
}
