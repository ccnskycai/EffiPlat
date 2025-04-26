# 数据库设计

## 1. 数据库选型

*   **初步选择**: PostgreSQL 或 MySQL
*   **理由**: 项目涉及大量实体和它们之间的复杂关联关系（用户、职责、环境、资产、服务、业务、配置、Bug 等），关系型数据库能很好地处理这些关系，并保证数据一致性。PostgreSQL 在复杂查询和扩展性方面略有优势，MySQL 则拥有广泛的社区支持和使用案例。最终选择需根据团队熟悉度和具体性能需求进一步评估。
*   **ORM/数据访问层**: 后端服务应使用 ORM 框架 (如 GORM for Go) 或轻量级库来简化数据库交互。

## 2. 核心实体与关系 (Conceptual Model - 初步)

[待细化: 使用 Mermaid ER 图或文字描述核心实体及其关系]

*   **用户 (Users)**: id, name, email, department, password_hash, status, created_at, updated_at
*   **角色 (Roles)**: id, name, description
*   **权限 (Permissions)**: id, resource, action (e.g., 'environment', 'read')
*   **用户角色关联 (UserRoles)**: user_id, role_id
*   **角色权限关联 (RolePermissions)**: role_id, permission_id
*   **职责 (Responsibilities)**: id, name, description
*   **职责组 (ResponsibilityGroups)**: id, responsibility_id, user_id, is_primary
*   **环境 (Environments)**: id, name, code, description, type, status, created_at, updated_at
*   **资产 (Assets)**: id, name, type (server, network_device), ... (根据类型细化字段)
*   **服务器资产 (ServerAssets)**: asset_id, ip_address, os, hostname, spec, access_info, ...
*   **服务 (Services)**: id, name, description, type, created_at, updated_at
*   **服务实例 (ServiceInstances)**: id, service_id, environment_id, server_asset_id, port, status, version
*   **业务 (Businesses)**: id, name, description, created_at, updated_at
*   **业务服务关联 (BusinessServices)**: business_id, service_id
*   **配置 (Configurations)**: id, service_id/business_id, environment_id, key, value, version, description, created_at, updated_at
*   **Bug 报告 (Bugs)**: id, title, description, status, priority, reporter_id, assignee_group_id, environment_id, service_instance_id, business_id, created_at, updated_at
*   **审计日志 (AuditLogs)**: id, user_id, timestamp, action, target_type, target_id, details (JSON/Text)

*关系示例*:*   一个用户可以有多个角色 (`UserRoles`)。
*   一个角色可以有多个权限 (`RolePermissions`)。
*   一个职责组 `ResponsibilityGroups` 包含一个或多个用户，并关联到一个职责 `Responsibilities`。
*   环境 `Environments` 可以关联多个服务实例 `ServiceInstances` 和资产 `Assets` (通过中间表或外键)。
*   服务实例 `ServiceInstances` 关联服务 `Services`、环境 `Environments` 和服务器资产 `ServerAssets`。
*   Bug `Bugs` 关联用户 (reporter), 职责组 (assignee), 环境, 服务实例, 业务等。

## 3. 核心表结构设计 (V1.0 关注点)

[根据 V1.0 计划 (`docs/requirements/execution_plan.md`)，优先设计以下核心表的详细结构]

*   `users`, `roles`, `user_roles` (支撑用户和基础权限)
*   `responsibilities`, `responsibility_groups` (支撑职责分配)
*   `environments`
*   `assets` (至少包含服务器类型的基础字段)
*   `services`
*   `service_instances`
*   `businesses`
*   `bugs` (包含核心关联字段)
*   `audit_logs` (基础结构)

[待补充: 每个核心表的详细字段定义、数据类型、约束、索引策略]

## 4. 数据一致性与完整性

*   使用外键约束强制实体间的关联关系。
*   关键字段（如 email, 服务名+环境）考虑添加唯一约束。
*   使用数据库事务保证操作的原子性，特别是在涉及多个表更新的操作（如创建 Bug 时关联多个实体）。

## 5. 数据迁移

*   如果现有系统存在相关数据，需制定数据迁移计划（见 `requirements.md` 6.8）。可能需要编写迁移脚本。 