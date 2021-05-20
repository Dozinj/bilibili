package param

type PostVideo struct {
	Title string `form:"title"`
	Length string `form:"length"`
	Channel string  `form:"channel"`
	Description string `form:"description"`
	Token string  `form:"token"`
}
