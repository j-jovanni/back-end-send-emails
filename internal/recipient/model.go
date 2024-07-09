package recipient

type Recipient struct {
	ID           string `bson:"_id,omitempty"`
	TemplateID   string `bson:"templateId"`
	Email        string `bson:"email"`
	Unsubscribed bool   `bson:"unsubscribed"`
}
