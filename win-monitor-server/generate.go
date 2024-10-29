package main

//需要先安装 swag 工具
// //go:generate  go install github.com/swaggo/swag/cmd/swag@latest
// //go:generate  go install github.com/go-swagger/go-swagger/cmd/swagger@v0.30.5
// 运行 swagger 服务，apiPost可以通过 http://localhost:4190/swagger.json 接口自动更新，不需要手动上传文件
// 生成 swagger 文档
//go:generate swag  fmt
//go:generate swag  init

//go:generate swagger serve ./docs/swagger.json /p 4190
// 更新接口到 apifox 文档
////go:generate go run ./script/apifox.go
