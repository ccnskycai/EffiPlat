# 审计日志实现指南

本文档提供了在EffiPlat中如何正确实现和使用审计日志功能的指导。

## 基本概念

审计日志记录了系统中的操作行为，包括：
- 谁（操作者）
- 什么时候（时间戳）
- 做了什么（操作类型）
- 针对什么资源（资源类型、资源ID）
- 具体细节（操作前后的状态变化等）

## 审计日志记录的实现方式

EffiPlat使用中间件和处理器相结合的方式记录审计日志：

1. `audit_log_middleware.go`中的`AuditLogMiddleware`自动捕获请求信息
2. 各处理器（handler）通过上下文（context）设置详细的审计信息

## 实现步骤

### 1. 在处理器中设置审计信息

每个处理器方法需要做以下工作：

```go
// 1. 设置审计操作类型、资源类型和资源ID
c.Set("auditAction", string(utils.AuditActionCreate)) // 或UPDATE, DELETE, READ
c.Set("auditResource", "USER") // 或其他资源类型，全大写
c.Set("auditResourceID", user.ID) // 资源ID，如果已知

// 2. 设置详细的审计信息
// 对于创建操作
auditDetails := utils.NewCreateAuditLog(map[string]interface{}{
    "id": entity.ID,
    "name": entity.Name,
    // 其他需要记录的字段
})

// 对于更新操作（记录前后变化）
beforeUpdate := map[string]interface{}{
    // 更新前的状态
}
afterUpdate := map[string]interface{}{
    // 更新后的状态
}
auditDetails := utils.NewUpdateAuditLog(beforeUpdate, afterUpdate)

// 对于删除操作
deleteDetails := map[string]interface{}{
    // 被删除的实体信息
}
auditDetails := utils.NewDeleteAuditLog(deleteDetails)

// 设置到上下文中
utils.SetAuditDetails(c, auditDetails)
```

### 2. 审计日志工具函数

系统提供了以下工具函数用于审计日志记录：

- `utils.SetAuditDetails(c, details)` - 设置审计详情
- `utils.NewCreateAuditLog(entity)` - 创建操作的审计日志
- `utils.NewUpdateAuditLog(before, after)` - 更新操作的审计日志
- `utils.NewDeleteAuditLog(entity)` - 删除操作的审计日志

### 3. 标准审计操作类型

使用`utils.AuditActionType`中定义的标准操作类型：

- `utils.AuditActionCreate` - 创建操作
- `utils.AuditActionRead` - 读取操作
- `utils.AuditActionUpdate` - 更新操作
- `utils.AuditActionDelete` - 删除操作
- `utils.AuditActionLogin` - 登录操作
- `utils.AuditActionLogout` - 登出操作

## 最佳实践

1. **不记录敏感信息**：不要在审计日志中记录密码、密钥等敏感信息
2. **记录资源变化**：对于更新操作，记录更新前后的状态变化
3. **适当粒度**：对于批量操作，可以记录操作的资源数量和类型，而不必记录每个资源的详细信息
4. **异常处理**：即使是失败的操作，也应考虑记录审计日志
5. **一致性**：保持审计日志记录方式的一致性，便于后续查询和分析

## 示例

### 创建操作的审计日志

```go
func (h *UserHandler) CreateUser(c *gin.Context) {
    // ... 请求解析和验证 ...
    
    // 设置审计日志信息
    c.Set("auditAction", string(utils.AuditActionCreate))
    c.Set("auditResource", "USER")
    
    // 执行创建操作
    user, err := h.userService.CreateUser(...)
    if err != nil {
        // ... 错误处理 ...
        return
    }
    
    // 设置详细的审计信息
    auditDetails := utils.NewCreateAuditLog(map[string]interface{}{
        "id": user.ID,
        "name": user.Name,
        "email": user.Email,
        // 不包含密码等敏感信息
    })
    utils.SetAuditDetails(c, auditDetails)
    c.Set("auditResourceID", user.ID)
    
    // ... 返回响应 ...
}
```

### 更新操作的审计日志

```go
func (h *RoleHandler) UpdateRole(c *gin.Context) {
    // ... 解析请求参数 ...
    
    // 设置审计日志信息
    c.Set("auditAction", string(utils.AuditActionUpdate))
    c.Set("auditResource", "ROLE")
    c.Set("auditResourceID", uint(roleID))
    
    // 获取更新前的状态
    existingRole, err := h.roleService.GetRoleByID(...)
    if err != nil {
        // ... 错误处理 ...
        return
    }
    
    // 记录更新前的状态
    beforeUpdate := map[string]interface{}{
        "id": existingRole.ID,
        "name": existingRole.Name,
        "description": existingRole.Description,
    }
    
    // 执行更新操作
    updatedRole, err := h.roleService.UpdateRole(...)
    if err != nil {
        // ... 错误处理 ...
        return
    }
    
    // 记录更新后的状态并设置审计日志
    afterUpdate := map[string]interface{}{
        "id": updatedRole.ID,
        "name": updatedRole.Name,
        "description": updatedRole.Description,
        "permissionIds": req.PermissionIDs,
    }
    
    auditDetails := utils.NewUpdateAuditLog(beforeUpdate, afterUpdate)
    utils.SetAuditDetails(c, auditDetails)
    
    // ... 返回响应 ...
}
```
