package gorm

import "sync"

// StmtCache 缓存中的 stmt 一定是有效的吗？
// 如何淘汰缓存中的 stmt ?
var StmtCache sync.Map