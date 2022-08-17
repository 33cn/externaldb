package main

import (
	_ "github.com/33cn/externaldb/rpc/docs"

	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func init() {
	swaggerHandler = ginSwagger.WrapHandler(swaggerFiles.Handler)
}
