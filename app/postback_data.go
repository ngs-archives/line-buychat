package app

// PostbackAction PostbackAction
type PostbackAction string

const (
	// PostbackActionAddCart add-cart
	PostbackActionAddCart PostbackAction = "add-cart"
)

// PostbackData PostbackData
type PostbackData struct {
	Action   PostbackAction
	ASIN     string
	ImageURL string
	Label    string
	Title    string
}
