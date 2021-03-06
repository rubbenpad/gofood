package domain

type Transaction struct {
	ID       string `json:"id,omitempty"`
	Device   string `json:"device,omitempty"`
	When     Uid    `json:"when,omitempty"`
	Products []Uid  `json:"products,omitempty"`
	From     Ip     `json:"from,omitempty"`
	Owner    Uid    `json:"owner,omitempty"`
}
