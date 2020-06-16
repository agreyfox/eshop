package main

import (
	"testing"

	"github.com/agreyfox/eshop/system/admin"
)

func TestSend(t *testing.T) {
	mail := admin.NewMailClient("grimmnanettehjbb@gmail.com", "qweasdzxC123^&*")
	mail.Send("标题1", "邮箱内容1", "jihua.gao@gmail.com") //邮件标题 邮件内容 需要发送到的邮箱地址
}
