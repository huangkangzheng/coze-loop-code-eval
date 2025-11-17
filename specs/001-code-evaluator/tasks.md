# Tasks: Code Evaluator

**Input**: Design documents from `/specs/001-code-evaluator/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are included in this plan as specified in plan.md Success Criteria (单元测试覆盖率 ≥ 80%, 集成测试全覆盖)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each backend API that supports the frontend user stories.

**Important**: This is a **backend-only** implementation. Frontend changes are handled by the frontend team. Tasks focus on the backend APIs that support the user stories defined in spec.md.

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3) - only for user story phases
- Include exact file paths in descriptions

## Path Conventions

This project uses **Web app (backend-only)** structure per plan.md:
- Backend: `backend/` (DDD architecture: application/, domain/, infra/)
- Sandbox: `sandbox/` (独立Docker服务)
- IDL: `backend/idl/`
- Migrations: `backend/cmd/sql_migration/migrations/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and code generation tools preparation

- [X] T001 Verify Go 1.21+ environment and all required dependencies (Kitex, Gorm, Wire, Viper)
- [X] T002 [P] Verify Deno 1.45.5 installation for sandbox development
- [X] T003 [P] Create sandbox directory structure: sandbox/Dockerfile, entrypoint.sh, healthcheck.sh, sandbox_server.ts

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete. These are blocking tasks for all subsequent work.

### IDL变更 (upgrade-idl skill)

- [ ] T004 Update backend/idl/evaluator.thrift - change LanguageType from enum to typedef string per prompt-contracts.md
- [ ] T005 [P] Update backend/idl/evaluator.thrift - modify CodeEvaluator struct (add code_template_key, code_template_name, rename code to code_content)
- [ ] T006 [P] Update backend/idl/evaluator.thrift - add mapstructure tags to CodeEvaluator fields
- [ ] T007 [P] Update backend/idl/evaluator.thrift - add EvaluatorContent.code_evaluator field with mapstructure tag
- [ ] T008 [P] Update backend/idl/evaluator.thrift - extend EvaluatorInputData (add evaluate_dataset_fields, evaluate_target_output_fields, ext)
- [ ] T009 [P] Update backend/idl/evaluator.thrift - extend EvaluatorOutputData (add stdout field)
- [ ] T010 Update backend/idl/coze/loop/evaluation/evaluator.thrift - add ValidateEvaluatorRequest/Response structs
- [ ] T011 [P] Update backend/idl/coze/loop/evaluation/evaluator.thrift - add BatchDebugEvaluatorRequest/Response structs
- [ ] T012 [P] Update backend/idl/coze/loop/evaluation/evaluator.thrift - extend GetTemplateInfoRequest (add optional language_type field)
- [ ] T013 Run kitex code generation for all modified IDL files (upgrade-idl skill will handle this)

### 数据库变更 (upgrade-sql skill)

- [ ] T014 Create migration script backend/cmd/sql_migration/migrations/YYYYMMDDHHMMSS_add_code_evaluator_fields.up.sql
- [ ] T015 [P] Add ALTER TABLE statement to add language_type VARCHAR(16) DEFAULT NULL to t_evaluator_version
- [ ] T016 [P] Add ALTER TABLE statement to add code_content TEXT DEFAULT NULL to t_evaluator_version
- [ ] T017 [P] Add ALTER TABLE statement to add code_template_key VARCHAR(64) DEFAULT NULL to t_evaluator_version
- [ ] T018 [P] Add ALTER TABLE statement to add code_template_name VARCHAR(128) DEFAULT NULL to t_evaluator_version
- [ ] T019 [P] Create index idx_code_template on (code_template_key, language_type, deleted_at)
- [ ] T020 Create rollback migration script backend/cmd/sql_migration/migrations/YYYYMMDDHHMMSS_add_code_evaluator_fields.down.sql
- [ ] T021 Run migration: go run backend/cmd/sql_migration/main.go up
- [ ] T022 Verify table structure: SHOW CREATE TABLE t_evaluator_version

### 沙箱服务开发 (独立Docker服务)

- [ ] T023 Create sandbox/Dockerfile with Deno 1.45.5 base image (copy exactly from prompt-sandbox.md lines 9-22)
- [ ] T024 [P] Create sandbox/entrypoint.sh with health check and server startup (copy exactly from prompt-sandbox.md lines 27-42)
- [ ] T025 [P] Create sandbox/healthcheck.sh with curl-based health check (copy exactly from prompt-sandbox.md lines 47-50)
- [ ] T026 Create sandbox/sandbox_server.ts - Part 1: Imports and types (lines 1-95 from prompt-sandbox.md)
- [ ] T027 [P] Create sandbox/sandbox_server.ts - Part 2: Python executor (lines 97-370 from prompt-sandbox.md)
- [ ] T028 [P] Create sandbox/sandbox_server.ts - Part 3: JavaScript executor (lines 372-562 from prompt-sandbox.md)
- [ ] T029 [P] Create sandbox/sandbox_server.ts - Part 4: HTTP server and routes (lines 564-740 from prompt-sandbox.md)
- [ ] T030 Build sandbox Docker image: docker build -t evaluator-sandbox:v1 sandbox/
- [ ] T031 Run sandbox container: docker run -d -p 8080:8080 --name sandbox evaluator-sandbox:v1
- [ ] T032 Verify sandbox health: curl http://localhost:8080/health

### 配置变更 (upgrade-config skill)

- [ ] T033 Update backend/conf/evaluation_config.go - add CodeEvaluatorTemplate struct with mapstructure tags
- [ ] T034 Update backend/conf/evaluation.yaml - add evaluator_template_conf.code.builtin_template_contains_any_python (from prompt-eval-tpl-conf.md lines 12-18)
- [ ] T035 [P] Update backend/conf/evaluation.yaml - add builtin_template_contains_any_js (lines 19-25)
- [ ] T036 [P] Update backend/conf/evaluation.yaml - add builtin_template_equal_python (lines 26-32)
- [ ] T037 [P] Update backend/conf/evaluation.yaml - add builtin_template_equal_js (lines 33-39)
- [ ] T038 [P] Update backend/conf/evaluation.yaml - add builtin_template_is_valid_json_object_python (lines 40-46)
- [ ] T039 [P] Update backend/conf/evaluation.yaml - add builtin_template_is_valid_json_object_js (lines 47-53)
- [ ] T040 [P] Update backend/conf/evaluation.yaml - add builtin_template_regex_python (lines 54-60)
- [ ] T041 [P] Update backend/conf/evaluation.yaml - add builtin_template_regex_js (lines 61-67)
- [ ] T042 [P] Update backend/conf/evaluation.yaml - add builtin_template_starts_with_python (lines 68-74)
- [ ] T043 [P] Update backend/conf/evaluation.yaml - add builtin_template_starts_with_js (lines 75-81)
- [ ] T044 [P] Update backend/conf/evaluation.yaml - add builtin_template_custom_python (lines 82-88)
- [ ] T045 [P] Update backend/conf/evaluation.yaml - add builtin_template_custom_js (lines 89-95)
- [ ] T046 Update backend/conf/evaluation.yaml - add evaluator_template_conf_en-US.code section with English template names

**Checkpoint**: Foundation ready - Domain layer development can now begin

---

## Phase 3: User Story 1+2+3 - Code评估器创建和测试 (Priority: P1) 🎯 MVP

**Goal**: 后端支持Code评估器的创建、配置、试运行和提交前代码检查。对应前端User Story 1(模板选择)、User Story 2(配置和测试)、User Story 3(提交前检查)。

**Backend APIs Required**:
- `GET /api/evaluation/v1/evaluators/templates/Code/{key}?language_type={lang}` - 获取模板信息
- `POST /api/evaluation/v1/evaluators/validate` - 代码验证和试运行
- `POST /api/evaluation/v1/evaluators` - 创建评估器（现有接口扩展）

**Independent Test**: 通过Postman/curl调用ValidateEvaluator接口验证Python/JS代码，调用GetTemplateInfo获取12个模板，调用CreateEvaluator创建Code类型评估器并验证数据库记录。

### Domain Layer Implementation for US1+2+3

- [ ] T047 [P] [US1] Create backend/domain/evaluator/entity/language_type.go - define LanguageType constants (Python, JS)
- [ ] T048 [P] [US1] Create backend/domain/evaluator/entity/code_evaluator_version.go - implement CodeEvaluatorVersion struct and IEvaluatorVersion interface
- [ ] T049 [P] [US2] Create backend/domain/evaluator/entity/sandbox_input.go - define SandboxExecutionInput struct
- [ ] T050 [P] [US2] Create backend/domain/evaluator/entity/sandbox_result.go - define SandboxExecutionResult and SandboxErrorType
- [ ] T051 [P] [US3] Create backend/domain/evaluator/entity/validation_result.go - define CodeValidationResult struct
- [ ] T052 [US2] Create backend/domain/evaluator/provider/isandbox_provider.go - define ISandboxProvider interface (Execute, HealthCheck methods)
- [ ] T053 [US2] Create backend/domain/evaluator/service/evaluator_source_code_service_impl.go - implement EvaluatorSourceCodeServiceImpl with Execute method calling sandboxProvider

### Infrastructure Layer Implementation for US1+2+3

- [ ] T054 [US2] Create backend/infra/provider/sandbox/sandbox_http_client.go - implement SandboxHTTPClient struct with Execute and HealthCheck methods
- [ ] T055 [US1] Update backend/infra/repo/evaluator/converter/evaluator_converter.go - add ConvertCodeEvaluatorVersionDO2PO function
- [ ] T056 [P] [US1] Update backend/infra/repo/evaluator/converter/evaluator_converter.go - add ConvertCodeEvaluatorVersionPO2DO function
- [ ] T057 [US1] Update backend/infra/repo/evaluator/evaluator_impl.go - fix BatchGetEvaluatorByVersionID switch to add case for EvaluatorTypeCode
- [ ] T058 [US1] Run gorm gen to regenerate backend/infra/db/model/evaluator_version.gen.go with new fields

### Application Layer Implementation for US1+2+3

- [ ] T059 [P] [US1] Update backend/application/evaluator/converter/evaluator_converter.go - add ConvertEvaluatorContentDTO2DO for Code type
- [ ] T060 [P] [US1] Update backend/application/evaluator/converter/evaluator_converter.go - add ConvertCodeEvaluatorVersionDO2DTO function
- [ ] T061 [P] [US2] Update backend/application/evaluator/converter/evaluator_converter.go - add ConvertSandboxResult2EvaluatorOutputData function
- [ ] T062 [US3] Implement ValidateEvaluator method in backend/application/evaluator/service/evaluator_service_impl.go
- [ ] T063 [US2] Implement BatchDebugEvaluator method in backend/application/evaluator/service/evaluator_service_impl.go (simple for-loop wrapper)
- [ ] T064 [US1] Update GetTemplateInfo method in backend/application/evaluator/service/evaluator_service_impl.go to support Code templates with language_type

### Wire依赖注入 (upgrade-wire skill)

- [ ] T065 Update backend/cmd/evaluator/wire.go - add sandbox.NewSandboxHTTPClient to wire.Build
- [ ] T066 [P] Update backend/cmd/evaluator/wire.go - add service.NewEvaluatorSourceCodeService to wire.Build
- [ ] T067 [P] Update backend/cmd/evaluator/main.go - add sandbox base URL to config (e.g., http://localhost:8080)
- [ ] T068 Run wire generate: cd backend/cmd/evaluator && wire

### Unit Tests for US1+2+3

- [ ] T069 [P] [US1] Create backend/infra/repo/evaluator/converter/evaluator_converter_test.go - test ConvertCodeEvaluatorVersionDO2PO (100% coverage)
- [ ] T070 [P] [US1] Create backend/infra/repo/evaluator/converter/evaluator_converter_test.go - test ConvertCodeEvaluatorVersionPO2DO (100% coverage)
- [ ] T071 [P] [US2] Create backend/domain/evaluator/service/evaluator_source_code_service_impl_test.go - test Execute method with mock sandbox
- [ ] T072 [P] [US2] Create backend/infra/provider/sandbox/sandbox_http_client_test.go - test Execute and HealthCheck with mock HTTP server
- [ ] T073 [P] [US3] Create backend/application/evaluator/service/evaluator_service_impl_test.go - test ValidateEvaluator with various code scenarios
- [ ] T074 [P] [US2] Create backend/application/evaluator/service/evaluator_service_impl_test.go - test BatchDebugEvaluator with multiple input data

### Integration Tests for US1+2+3

- [ ] T075 [US1] Create backend/tests/integration/test_get_template_info.go - test GetTemplateInfo for all 12 Code templates (Python + JS)
- [ ] T076 [US3] Create backend/tests/integration/test_validate_evaluator.go - test ValidateEvaluator with valid/invalid Python code
- [ ] T077 [P] [US3] Create backend/tests/integration/test_validate_evaluator.go - test ValidateEvaluator with valid/invalid JS code
- [ ] T078 [US2] Create backend/tests/integration/test_batch_debug_evaluator.go - test BatchDebugEvaluator with 10 test inputs
- [ ] T079 [US1] Create backend/tests/integration/test_create_code_evaluator.go - test creating Code evaluator via existing CreateEvaluator API

### Compilation Verification

- [ ] T080 Compile backend: cd backend && sh cmd/build.sh
- [ ] T081 Fix any compilation errors and re-run build

**Checkpoint**: MVP完成 - Code评估器可以被创建、验证和调试。GetTemplateInfo返回12个模板，ValidateEvaluator和BatchDebugEvaluator正常工作。

---

## Phase 4: User Story 4 - 编辑过程中切换模板 (Priority: P2)

**Goal**: 后端无需额外API支持。前端通过重新调用GetTemplateInfo并更新表单即可实现。

**Backend APIs Required**: 无（复用Phase 3的GetTemplateInfo接口）

**Independent Test**: 前端功能，后端无需独立测试。

**Tasks**: 无后端任务（This is a frontend-only user story, no backend changes required）

**Checkpoint**: US4无后端工作，依赖US1的GetTemplateInfo接口。

---

## Phase 5: User Story 5 - 查看和管理Code评估器 (Priority: P2)

**Goal**: 后端支持评估器列表返回类型字段，以便前端展示和筛选。

**Backend APIs Required**:
- `GET /api/evaluation/v1/evaluators` - 现有接口，需要在响应中包含evaluator_type字段

**Independent Test**: 调用ListEvaluators接口，验证返回的评估器列表包含Code和Prompt两种类型，type字段正确。

### Implementation for US5

- [ ] T082 [US5] Verify backend/application/evaluator/service/evaluator_service_impl.go ListEvaluators method includes evaluator_type in response DTO
- [ ] T083 [US5] Update ListEvaluators DTO converter if needed to ensure type field is populated from entity.EvaluatorType

### Integration Tests for US5

- [ ] T084 [US5] Create backend/tests/integration/test_list_evaluators.go - test ListEvaluators returns both Code and Prompt types with correct type field

**Checkpoint**: 评估器列表接口返回类型字段，前端可以展示"Code"或"LLM"标识。

---

## Phase 6: User Story 6 - 在实验中使用Code评估器 (Priority: P2)

**Goal**: 后端支持实验运行时调用Code评估器执行评估，并返回结果。

**Backend APIs Required**:
- `GET /api/evaluation/v1/evaluators/versions/{id}` - 修复后支持Code类型（已在Phase 3完成T057）
- Experiment execution logic - 调用Code评估器的Execute方法（可能需要修改实验执行模块）

**Independent Test**: 创建包含Code评估器的实验，运行实验，验证评估结果正确输出到实验结果中。

### Implementation for US6

- [ ] T085 [US6] Verify experiment execution module backend/application/experiment/service/experiment_runner.go can handle EvaluatorTypeCode
- [ ] T086 [US6] Update experiment evaluation logic to call EvaluatorSourceCodeService.Execute for Code evaluators
- [ ] T087 [US6] Ensure experiment results include evaluator_result.score and evaluator_result.reason for Code evaluators

### Integration Tests for US6

- [ ] T088 [US6] Create backend/tests/integration/test_experiment_with_code_evaluator.go - test running experiment with Code evaluator
- [ ] T089 [US6] Verify experiment results contain correct score and reason from Code evaluator execution

**Checkpoint**: 实验可以使用Code评估器进行评估，结果正确记录。

---

## Phase 7: User Story 7 - 查看Code评估器详情和试运行 (Priority: P3)

**Goal**: 后端支持详情页的试运行功能（复用BatchDebugEvaluator接口）和详情查询。

**Backend APIs Required**:
- `GET /api/evaluation/v1/evaluators/{id}` - 现有接口，需确保返回Code类型评估器的完整信息
- `POST /api/evaluation/v1/evaluators/batch_debug` - 已在Phase 3实现（T063）

**Independent Test**: 调用GetEvaluator接口获取Code评估器详情，调用BatchDebugEvaluator在详情页试运行。

### Implementation for US7

- [ ] T090 [US7] Verify backend/application/evaluator/service/evaluator_service_impl.go GetEvaluator method returns Code evaluator details correctly
- [ ] T091 [US7] Ensure GetEvaluator DTO includes code_evaluator field with all necessary information (language_type, code_content, template_key, template_name)

### Integration Tests for US7

- [ ] T092 [US7] Create backend/tests/integration/test_get_code_evaluator_detail.go - test GetEvaluator returns complete Code evaluator details
- [ ] T093 [US7] Test BatchDebugEvaluator can be called from evaluator detail page context

**Checkpoint**: Code评估器详情页数据完整，试运行功能复用BatchDebugEvaluator接口。

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Performance optimization, error handling enhancement, documentation, and final validation

### Performance & Error Handling

- [ ] T094 [P] Add timeout handling (30s) for sandbox HTTP calls in backend/infra/provider/sandbox/sandbox_http_client.go
- [ ] T095 [P] Add retry logic for sandbox health check failures in sandbox HTTP client
- [ ] T096 [P] Add comprehensive error logging for sandbox execution failures with stack traces
- [ ] T097 [P] Implement sandbox connection pool if needed for concurrent requests
- [ ] T098 [P] Add monitoring metrics for sandbox execution time (P95, P99)

### Security & Validation

- [ ] T099 [P] Validate code_content size limit (100KB) in ValidateEvaluator
- [ ] T100 [P] Validate input_data size limit (1MB) in BatchDebugEvaluator
- [ ] T101 [P] Validate batch size limit (≤10条) in BatchDebugEvaluator
- [ ] T102 [P] Add workspace_id validation in all Code evaluator APIs

### Documentation

- [ ] T103 [P] Add API documentation for ValidateEvaluator endpoint with request/response examples
- [ ] T104 [P] Add API documentation for BatchDebugEvaluator endpoint with request/response examples
- [ ] T105 [P] Update backend README with sandbox deployment instructions
- [ ] T106 [P] Document supported Python libraries and JavaScript APIs in backend/docs/code_evaluator_libraries.md

### Final Validation

- [ ] T107 Run full test suite: go test ./... -v -cover
- [ ] T108 Verify test coverage ≥ 80% for Domain Service and Application Service layers
- [ ] T109 Verify Converter layer test coverage = 100%
- [ ] T110 Run integration tests against local MySQL and sandbox
- [ ] T111 Verify all 4 core API endpoints (GetTemplateInfo, ValidateEvaluator, BatchDebugEvaluator, GetEvaluatorVersion) pass integration tests
- [ ] T112 Performance test: Verify ValidateEvaluator P95 < 5s with test execution
- [ ] T113 Performance test: Verify BatchDebugEvaluator P95 < 30s with 10 inputs
- [ ] T114 Final compilation check: cd backend && sh cmd/build.sh

**Checkpoint**: All functionality complete, tested, and validated. Ready for deployment.

---

## Task Dependencies & Execution Strategy

### Critical Path (Must complete in order)

1. **Phase 1** (Setup) → **Phase 2** (Foundational) → All other phases
2. Within Phase 2: IDL变更 (T004-T013) → 数据库变更 (T014-T022) → 配置变更 (T033-T046)
3. Phase 2 沙箱服务 (T023-T032) can run in parallel with IDL/DB/配置变更
4. **Phase 3** (US1+2+3 MVP) blocks Phase 6 and Phase 7
5. Phase 4, Phase 5 can run in parallel with Phase 6, Phase 7 after Phase 3 completes

### Parallel Execution Opportunities

**Phase 2 Parallelization** (after T004-T013 IDL generation completes):
- T023-T032 (沙箱服务) || T014-T022 (数据库) || T033-T046 (配置)

**Phase 3 Parallelization** (within Domain/Infra/Application layers):
- T047-T051 (Domain entities) can all run in parallel
- T055-T056 (Converters) can run in parallel
- T059-T061 (Application converters) can run in parallel
- T069-T074 (Unit tests) can all run in parallel after implementation tasks
- T075-T079 (Integration tests) can run in parallel after T080 compilation succeeds

**Phase 8 Parallelization**:
- T094-T102 (所有优化任务) can run in parallel
- T103-T106 (所有文档任务) can run in parallel

### Suggested MVP Scope (Minimum Deliverable)

**MVP = Phase 1 + Phase 2 + Phase 3 ONLY**

This delivers:
- ✅ 沙箱服务部署和运行
- ✅ 12个Code评估器模板（Python + JS）
- ✅ ValidateEvaluator接口（代码验证 + 试运行）
- ✅ BatchDebugEvaluator接口（批量调试）
- ✅ GetTemplateInfo接口支持Code类型
- ✅ CreateEvaluator接口支持Code类型
- ✅ GetEvaluatorVersion接口修复支持Code类型

MVP allows users to:
1. 选择模板创建Code评估器
2. 配置和测试执行函数
3. 提交前进行代码检查
4. 创建Code评估器成功

**Post-MVP** (Phase 4-8) adds:
- US4: 模板切换（前端功能，无后端工作）
- US5: 列表类型展示
- US6: 实验中使用Code评估器
- US7: 详情页试运行
- Phase 8: 性能优化、安全加固、文档完善

---

## Total Task Count

- **Phase 1** (Setup): 3 tasks
- **Phase 2** (Foundational): 43 tasks
- **Phase 3** (US1+2+3 MVP): 35 tasks
- **Phase 4** (US4): 0 tasks
- **Phase 5** (US5): 3 tasks
- **Phase 6** (US6): 5 tasks
- **Phase 7** (US7): 4 tasks
- **Phase 8** (Polish): 21 tasks

**Total**: 114 tasks

**MVP Tasks** (Phase 1+2+3): 81 tasks
**Post-MVP Tasks** (Phase 4-8): 33 tasks

**Parallelization**: ~60% of tasks can run in parallel (marked with [P] or within parallelizable groups)

**Estimated Effort**:
- MVP: ~20-24 hours (3个工作日)
- Full Feature: ~30 hours (4个工作日)

---

## Validation Checklist

### Format Validation ✅

All tasks follow the required format:
- [x] All tasks have checkbox prefix `- [ ]`
- [x] All tasks have sequential IDs (T001-T114)
- [x] User story phase tasks have [US#] labels
- [x] Setup and Foundational phases have NO story labels
- [x] Polish phase has NO story labels
- [x] Parallelizable tasks marked with [P]
- [x] All tasks include exact file paths or clear descriptions

### Content Validation ✅

- [x] Tasks organized by user story (US1+2+3 combined as single MVP phase)
- [x] Each user story phase has Goal and Independent Test criteria
- [x] Dependencies clearly documented (Critical Path section)
- [x] Parallel execution opportunities identified
- [x] MVP scope clearly defined (Phase 1+2+3)
- [x] Test tasks included per plan.md requirements (单元测试、集成测试)
- [x] Backend-only focus maintained (frontend changes excluded)

### Technical Validation ✅

- [x] All entities from data-model.md mapped to tasks
- [x] All API endpoints from contracts/ mapped to tasks
- [x] All user stories from spec.md have corresponding backend tasks
- [x] DDD architecture layers (Domain → Infra → Application) respected
- [x] Code generation workflows (upgrade-idl, upgrade-sql, upgrade-config, upgrade-wire) included
- [x] Sandbox implementation based on prompt-sandbox.md exact code
- [x] 12 templates from prompt-eval-tpl-conf.md all included

---

**Tasks Generated**: 2025-11-14 by `/speckit.tasks` command
**Status**: Ready for `/speckit.implement` execution
**Next Step**: Run `/speckit.implement` to begin Phase 1 execution
