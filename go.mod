module github.com/WANGgbin/mini_gorm

go 1.16

require (
	github.com/WANGgbin/mini_mysql_driver v0.0.0-20230918040803-781451ed3b68
	github.com/smartystreets/goconvey v1.7.0
	google.golang.org/protobuf v1.31.0 // indirect
	gorm.io/driver/mysql v1.5.1
	gorm.io/gorm v1.25.2
)

replace github.com/WANGgbin/mini_mysql_driver => ../mini_mysql_driver
