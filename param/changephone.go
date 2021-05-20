package param

type Param struct {
	Old_account string`form:"old_account"`
	Old_code string`form:"old_code"`
	New_phone string `form:"new_phone"`
	New_code string `form:"new_code"`
	Token string `form:"token"`
}

