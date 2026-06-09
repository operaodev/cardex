package transactions

type TransactionType string

const (
	SaleItemType  TransactionType = "sale"
	TradeItemType TransactionType = "trade"
	GiftItemType  TransactionType = "gift"
)

type Transaction struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	SenderID   uint64 `gorm:"not null;index" json:"sender_id"`
	ReceiverID uint64 `gorm:"not null;index" json:"receiver_id"`
	MediatorID uint64 `gorm:"" json:"mediator_id"`
}

type TransactionParticipant struct {
	ID      uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	OwnerID uint64 `gorm:"not null;index" json:"owner_id"`
}

type TransactionItem struct {
	ID              uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	InventoryID     uint64          `gorm:"not null;index" json:"inventory_id"`
	Quantity        uint64          `gorm:"not null" json:"quantity"`
	TransactionType TransactionType `gorm:"not null" json:"transaction_type"`
}
