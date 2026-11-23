package models

type Permission string

const (
	UserReadOwn Permission = "user.read.own"
	UserReadAll Permission = "user.read.all"
	UserUpdate  Permission = "user.update"
	UserDelete  Permission = "user.delete"
	UserSearch  Permission = "user.search"
	UserStats   Permission = "user.stats"

	AccountReadOwn Permission = "account.read.own"
	AccountReadAll Permission = "account.read.all"
	AccountUpdate  Permission = "account.update"
	AccountDelete  Permission = "account.delete"
	AccountSearch  Permission = "account.search"
	AccountStats   Permission = "account.stats"
)
