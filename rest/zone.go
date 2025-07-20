package rest

type Zone struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Serial  int64  `json:"serial"`
	Admin   string `json:"admin"`
	Refresh int32  `json:"refresh"`
	Retry   int32  `json:"retry"`
	Expire  int32  `json:"expire"`
	Ttl     int32  `json:"ttl"`
	Active  bool   `json:"active"`
}
