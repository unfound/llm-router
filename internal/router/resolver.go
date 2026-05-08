package router

// ResolveModel 别名解析 - 将模型别名解析为真实配置
// TODO: 从数据库查询模型配置
func ResolveModel(alias string) (provider, modelID, apiBase, apiKey string, err error) {
	// 临时实现：从配置文件读取
	return "", "", "", "", nil
}
