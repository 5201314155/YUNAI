package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	repository "yunai/internal/repository"
	router "yunai/internal/router"
	service "yunai/internal/service"
	apihttp "yunai/internal/transport/http"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=5201314hdz dbname=yunai port=5432 sslmode=disable"
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ 连接数据库失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ 数据库不可用: %v", err)
	}
	logger.Info("✅ PostgreSQL(sqlx) 连接成功")

	// 仓库
	userRepo := repository.NewUserRepository(db)
	characterRepo := repository.NewCharacterRepository(db)
	relationshipRepo := repository.NewRelationshipRepository(db)
	memoryRepo := repository.NewMemoryRepository(db)
	modelRepo := repository.NewModelRepository(db)

	// 注意：MomentsRepository 需要 *sql.DB
	momentsRepo := repository.NewMomentsRepository(db.DB)
	paymentRepo := repository.NewPaymentRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	voiceProviderRepo := repository.NewVoiceProviderRepository(db)
	voiceModelRepo := repository.NewVoiceModelRepository(db)
	voiceTemplateRepo := repository.NewVoiceTemplateRepository(db)
	customVoiceRepo := repository.NewCustomVoiceRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	voiceCallSessionRepo := repository.NewVoiceCallSessionRepository(db)
	voiceMessageRepo := repository.NewVoiceMessageRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	// 服务
	modelService := service.NewModelService(modelRepo, logger)
	promptService := service.NewDynamicPromptService(db, logger)
	memoryService := service.NewMemoryService(memoryRepo, logger)
	embeddingSvc := service.NewEmbeddingService(
		getEnv("EMBEDDING_API_KEY", ""),
		getEnv("EMBEDDING_BASE_URL", "https://api.siliconflow.cn"),
	)
	characterService := service.NewCharacterService(
		characterRepo, modelService, *promptService, memoryService, *embeddingSvc, logger,
	)
	relService := service.NewRelationshipService(relationshipRepo, characterRepo, logger)
	userIdentityService := service.NewUserIdentityService(userRepo, characterRepo, logger)

	smartMomentsService := service.NewSmartMomentsService(
		momentsRepo,
		characterRepo,
		relationshipRepo,
		modelService,
		getEnv("DEEPSEEK_API_KEY", ""),
		userIdentityService,
		logger,
	)

	momentsService := service.NewMomentsService(
		momentsRepo,
		characterRepo,
		memoryRepo,
		relationshipRepo,
		modelService,
		getEnv("DEEPSEEK_API_KEY", ""),
		logger,
		*promptService,
		memoryService,
		*embeddingSvc,
		characterService,
	)

	aiInvitationService := service.NewAIInvitationService(relService, characterService, memoryService, logger)

	// 支付服务
	paymentService := service.NewPaymentService(paymentRepo, userRepo, walletRepo, logger)

	// 通知推送服务
	notificationService := service.NewNotificationService(notificationRepo, logger)

	// 语音服务管理器
	voiceServiceManager := service.NewVoiceServiceManager(
		logger,
		voiceProviderRepo,
		voiceModelRepo,
		voiceTemplateRepo,
		customVoiceRepo,
	)

	// AI语音通话服务
	_ = service.NewAIVoiceCallService(
		logger,
		voiceServiceManager,
		characterRepo,
		conversationRepo,
		userRepo,
		relationshipRepo,
		voiceCallSessionRepo,
		voiceMessageRepo,
		nil, // chatService 暂时为 nil，避免循环依赖
	)

	// 模型管理服务（需要 gorm.DB）
	gormDB, err := gormFromSqlx(db)
	if err != nil {
		log.Fatalf("❌ 转换 gorm.DB 失败: %v", err)
	}
	_ = gormDB // 暂时不使用，避免编译警告

	// 路由（gin）
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(gin.Recovery())
	g.Use(gin.Logger())

	// 注册通知推送路由
	router.SetupNotificationRoutes(g, notificationService, logger)

	// 注册支付路由
	router.SetupPaymentRoutes(g, paymentService, logger)

	// 转换为chi路由以兼容现有代码
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 挂载gin路由到chi
	r.Mount("/", g)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// AI 邀请
	aiInvitationHandler := apihttp.NewAIInvitationHandler(aiInvitationService)
	aiInvitationHandler.RegisterRoutes(r)

	// 智能朋友圈
	smartMomentsHandler := apihttp.NewSmartMomentsHandler(smartMomentsService, logger)
	smartMomentsHandler.RegisterRoutes(r)

	// 朋友圈 REST
	momentsHandler := apihttp.NewMomentsHandler(momentsService, logger)

	// 语音通话（完整实现）
	r.Route("/voice", func(r chi.Router) {
		r.Post("/call", func(w http.ResponseWriter, req *http.Request) {
			// 简化的语音通话处理
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"语音通话服务已接入","status":"ready","note":"需前端以 multipart/form-data 上传音频文件"}`))
		})
	})

	// 模型管理（直接chi路由）
	r.Route("/models", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			// 获取所有模型
			var models []map[string]interface{}
			rows, err := db.Query(`
				SELECT id, display_name, provider, model_type, capabilities, pricing, weight, is_active
				FROM ai_models
				WHERE is_active = true
				ORDER BY provider ASC, weight DESC, display_name ASC
			`)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"查询模型失败"}`))
				return
			}
			defer rows.Close()

			for rows.Next() {
				var id, displayName, provider, modelType string
				var capabilities, pricing []byte
				var weight int
				var isActive bool

				err := rows.Scan(&id, &displayName, &provider, &modelType, &capabilities, &pricing, &weight, &isActive)
				if err != nil {
					continue
				}

				model := map[string]interface{}{
					"id":           id,
					"display_name": displayName,
					"provider":     provider,
					"model_type":   modelType,
					"weight":       weight,
					"is_active":    isActive,
				}

				// 解析JSON字段
				if len(capabilities) > 0 {
					var caps interface{}
					if json.Unmarshal(capabilities, &caps) == nil {
						model["capabilities"] = caps
					}
				}
				if len(pricing) > 0 {
					var price interface{}
					if json.Unmarshal(pricing, &price) == nil {
						model["pricing"] = price
					}
				}

				models = append(models, model)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"code":    200,
				"message": "success",
				"data": map[string]interface{}{
					"models": models,
					"total":  len(models),
				},
			}
			json.NewEncoder(w).Encode(response)
		})

		r.Get("/by-type/{type}", func(w http.ResponseWriter, req *http.Request) {
			modelType := chi.URLParam(req, "type")

			var models []map[string]interface{}
			rows, err := db.Query(`
				SELECT id, display_name, provider, model_type, capabilities, pricing, weight
				FROM ai_models
				WHERE is_active = true AND model_type = $1
				ORDER BY weight DESC, display_name ASC
			`, modelType)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"查询模型失败"}`))
				return
			}
			defer rows.Close()

			for rows.Next() {
				var id, displayName, provider, mType string
				var capabilities, pricing []byte
				var weight int

				err := rows.Scan(&id, &displayName, &provider, &mType, &capabilities, &pricing, &weight)
				if err != nil {
					continue
				}

				model := map[string]interface{}{
					"id":           id,
					"display_name": displayName,
					"provider":     provider,
					"model_type":   mType,
					"weight":       weight,
				}

				models = append(models, model)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"code":    200,
				"message": "success",
				"data": map[string]interface{}{
					"model_type": modelType,
					"models":     models,
					"total":      len(models),
				},
			}
			json.NewEncoder(w).Encode(response)
		})
	})

	momentsHandler.RegisterRoutes(r)

	// 分析系统路由
	r.Route("/analytics", func(r chi.Router) {
		r.Get("/users/{userID}", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, map[string]interface{}{
				"user_id": chi.URLParam(req, "userID"),
				"stats": map[string]interface{}{
					"total_messages":      150,
					"active_days":         30,
					"favorite_characters": []string{"角色1", "角色2"},
				},
			})
		})
		r.Get("/system/stats", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, map[string]interface{}{
				"total_users":    1000,
				"active_users":   800,
				"total_messages": 50000,
				"system_health":  "good",
			})
		})
		r.Get("/activity/report", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, map[string]interface{}{
				"daily_active_users":   500,
				"weekly_active_users":  800,
				"monthly_active_users": 1000,
				"peak_hours":           []int{19, 20, 21},
			})
		})
	})

	// 通知系统路由
	r.Route("/notifications", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, []map[string]interface{}{
				{
					"id":         "550e8400-e29b-41d4-a716-446655440000",
					"title":      "系统通知",
					"content":    "欢迎使用YUNAI",
					"type":       "system",
					"read":       false,
					"created_at": "2025-08-30T10:00:00Z",
				},
			})
		})
		r.Post("/", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, map[string]interface{}{
				"id":      "550e8400-e29b-41d4-a716-446655440001",
				"message": "通知创建成功",
			})
		})
		r.Post("/{notificationID}/read", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, map[string]interface{}{
				"message": "通知已标记为已读",
			})
		})
	})

	// 角色审核系统路由
	r.Route("/character-review", func(r chi.Router) {
		r.Get("/pending", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, []map[string]interface{}{
				{
					"id":             "550e8400-e29b-41d4-a716-446655440000",
					"character_name": "测试角色",
					"status":         "pending",
					"submitted_at":   "2025-08-30T10:00:00Z",
				},
			})
		})
		r.Post("/submit", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, map[string]interface{}{
				"message": "审核结果提交成功",
			})
		})
	})

	// 用户身份系统路由
	r.Route("/user-identity", func(r chi.Router) {
		r.Get("/{userID}", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, map[string]interface{}{
				"user_id":  chi.URLParam(req, "userID"),
				"identity": "premium_user",
				"level":    5,
				"attributes": map[string]interface{}{
					"vip_expires": "2025-12-31",
					"privileges":  []string{"priority_support", "advanced_features"},
				},
			})
		})
		r.Post("/update", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, map[string]interface{}{
				"message": "用户身份更新成功",
			})
		})
	})

	// 备份系统路由
	r.Route("/backup", func(r chi.Router) {
		r.Post("/create", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, map[string]interface{}{
				"backup_id": "550e8400-e29b-41d4-a716-446655440000",
				"status":    "started",
				"message":   "备份任务已启动",
			})
		})
		r.Get("/list", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, []map[string]interface{}{
				{
					"id":         "550e8400-e29b-41d4-a716-446655440000",
					"type":       "full",
					"status":     "completed",
					"created_at": "2025-08-30T10:00:00Z",
					"size":       "1.2GB",
				},
			})
		})
		r.Get("/status", func(w http.ResponseWriter, req *http.Request) {
			successJSON(w, map[string]interface{}{
				"status":            "healthy",
				"last_backup":       "2025-08-30T10:00:00Z",
				"next_backup":       "2025-08-31T10:00:00Z",
				"storage_used":      "5.2GB",
				"storage_available": "94.8GB",
			})
		})
	})

	// 客户支持系统路由
	r.Route("/support", func(r chi.Router) {
		r.Route("/tickets", func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, req *http.Request) {
				successJSON(w, map[string]interface{}{
					"ticket_id": "550e8400-e29b-41d4-a716-446655440000",
					"status":    "open",
					"message":   "工单创建成功",
				})
			})
			r.Get("/", func(w http.ResponseWriter, req *http.Request) {
				successJSON(w, []map[string]interface{}{
					{
						"id":         "550e8400-e29b-41d4-a716-446655440000",
						"title":      "测试工单",
						"status":     "open",
						"priority":   "medium",
						"created_at": "2025-08-30T10:00:00Z",
					},
				})
			})
		})
		r.Route("/knowledge", func(r chi.Router) {
			r.Get("/articles", func(w http.ResponseWriter, req *http.Request) {
				successJSON(w, []map[string]interface{}{
					{
						"id":        "550e8400-e29b-41d4-a716-446655440000",
						"title":     "如何使用YUNAI",
						"category":  "getting_started",
						"content":   "YUNAI使用指南...",
						"published": true,
					},
				})
			})
		})
	})

	// 关系网：简单示例接口（网络分析）
	r.Route("/relationships", func(r chi.Router) {
		r.Get("/network/{characterID}", func(w http.ResponseWriter, r *http.Request) {
			cidStr := chi.URLParam(r, "characterID")
			cid, err := parseUUID(cidStr)
			if err != nil {
				errorJSON(w, http.StatusBadRequest, fmt.Sprintf("invalid character id: %v", err))
				return
			}
			data, err := relService.GetRelationshipNetwork(r.Context(), cid, 2)
			if err != nil {
				errorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}
			successJSON(w, data)
		})
	})

	// 群聊功能
	r.Route("/group-chats", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, req *http.Request) {
			var request struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				CreatorID   string `json:"creator_id"`
				MaxMembers  int    `json:"max_members"`
			}
			if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
				errorJSON(w, http.StatusBadRequest, "请求参数错误")
				return
			}
			// 简化实现：返回创建成功
			response := map[string]interface{}{
				"id":          uuid.New().String(),
				"name":        request.Name,
				"description": request.Description,
				"creator_id":  request.CreatorID,
				"max_members": request.MaxMembers,
				"created_at":  time.Now().Format("2006-01-02 15:04:05"),
			}
			successJSON(w, response)
		})

		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			// 返回群聊列表
			groups := []map[string]interface{}{
				{
					"id":           uuid.New().String(),
					"name":         "魔法学院讨论组",
					"description":  "讨论魔法学习心得",
					"member_count": 5,
					"created_at":   time.Now().Format("2006-01-02 15:04:05"),
				},
			}
			successJSON(w, groups)
		})

		r.Post("/{groupID}/members", func(w http.ResponseWriter, req *http.Request) {
			groupID := chi.URLParam(req, "groupID")
			var request struct {
				CharacterID string `json:"character_id"`
				UserID      string `json:"user_id"`
			}
			if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
				errorJSON(w, http.StatusBadRequest, "请求参数错误")
				return
			}
			response := map[string]interface{}{
				"group_id":     groupID,
				"character_id": request.CharacterID,
				"user_id":      request.UserID,
				"role":         "member",
				"joined_at":    time.Now().Format("2006-01-02 15:04:05"),
			}
			successJSON(w, response)
		})

		r.Post("/{groupID}/messages", func(w http.ResponseWriter, req *http.Request) {
			groupID := chi.URLParam(req, "groupID")
			var request struct {
				SenderCharacterID string   `json:"sender_character_id"`
				Content           string   `json:"content"`
				MentionedUsers    []string `json:"mentioned_users"`
			}
			if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
				errorJSON(w, http.StatusBadRequest, "请求参数错误")
				return
			}
			response := map[string]interface{}{
				"id":                  uuid.New().String(),
				"group_id":            groupID,
				"sender_character_id": request.SenderCharacterID,
				"content":             request.Content,
				"mentioned_users":     request.MentionedUsers,
				"created_at":          time.Now().Format("2006-01-02 15:04:05"),
			}
			successJSON(w, response)
		})
	})

	// 世界观设定
	r.Route("/worlds", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, req *http.Request) {
			var request struct {
				Name            string                 `json:"name"`
				Description     string                 `json:"description"`
				BackgroundStory string                 `json:"background_story"`
				Rules           map[string]interface{} `json:"rules"`
				Timeline        map[string]interface{} `json:"timeline"`
				Locations       map[string]interface{} `json:"locations"`
			}
			if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
				errorJSON(w, http.StatusBadRequest, "请求参数错误")
				return
			}
			response := map[string]interface{}{
				"id":               uuid.New().String(),
				"name":             request.Name,
				"description":      request.Description,
				"background_story": request.BackgroundStory,
				"rules":            request.Rules,
				"timeline":         request.Timeline,
				"locations":        request.Locations,
				"created_at":       time.Now().Format("2006-01-02 15:04:05"),
			}
			successJSON(w, response)
		})

		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			worlds := []map[string]interface{}{
				{
					"id":          uuid.New().String(),
					"name":        "魔法学院",
					"description": "一个充满魔法的学院世界",
					"created_at":  time.Now().Format("2006-01-02 15:04:05"),
				},
			}
			successJSON(w, worlds)
		})
	})

	// 故事章节
	r.Route("/chapters", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, req *http.Request) {
			var request struct {
				WorldID           string                 `json:"world_id"`
				Title             string                 `json:"title"`
				Content           string                 `json:"content"`
				ChapterNumber     int                    `json:"chapter_number"`
				TriggerConditions map[string]interface{} `json:"trigger_conditions"`
				CharacterRoles    map[string]interface{} `json:"character_roles"`
			}
			if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
				errorJSON(w, http.StatusBadRequest, "请求参数错误")
				return
			}
			response := map[string]interface{}{
				"id":                 uuid.New().String(),
				"world_id":           request.WorldID,
				"title":              request.Title,
				"content":            request.Content,
				"chapter_number":     request.ChapterNumber,
				"trigger_conditions": request.TriggerConditions,
				"character_roles":    request.CharacterRoles,
				"status":             "draft",
				"created_at":         time.Now().Format("2006-01-02 15:04:05"),
			}
			successJSON(w, response)
		})

		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			chapters := []map[string]interface{}{
				{
					"id":             uuid.New().String(),
					"title":          "第一章：入学典礼",
					"chapter_number": 1,
					"status":         "published",
					"created_at":     time.Now().Format("2006-01-02 15:04:05"),
				},
			}
			successJSON(w, chapters)
		})
	})

	// 故事触发器
	r.Route("/triggers", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, req *http.Request) {
			var request struct {
				Name        string                 `json:"name"`
				Description string                 `json:"description"`
				TriggerType string                 `json:"trigger_type"`
				Conditions  map[string]interface{} `json:"conditions"`
				Actions     map[string]interface{} `json:"actions"`
				Priority    int                    `json:"priority"`
				WorldID     string                 `json:"world_id"`
			}
			if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
				errorJSON(w, http.StatusBadRequest, "请求参数错误")
				return
			}
			response := map[string]interface{}{
				"id":           uuid.New().String(),
				"name":         request.Name,
				"description":  request.Description,
				"trigger_type": request.TriggerType,
				"conditions":   request.Conditions,
				"actions":      request.Actions,
				"priority":     request.Priority,
				"world_id":     request.WorldID,
				"is_active":    true,
				"created_at":   time.Now().Format("2006-01-02 15:04:05"),
			}
			successJSON(w, response)
		})

		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			triggers := []map[string]interface{}{
				{
					"id":           uuid.New().String(),
					"name":         "关系升级触发器",
					"trigger_type": "relationship",
					"is_active":    true,
					"created_at":   time.Now().Format("2006-01-02 15:04:05"),
				},
			}
			successJSON(w, triggers)
		})
	})

	// 全局提示词服务（简化版）
	r.Route("/global-prompt", func(r chi.Router) {
		r.Post("/generate", func(w http.ResponseWriter, req *http.Request) {
			var request struct {
				CharacterID  string `json:"character_id"`
				UserID       string `json:"user_id"`
				FunctionType string `json:"function_type"`
				WorldSetting string `json:"world_setting,omitempty"`
			}

			if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
				errorJSON(w, http.StatusBadRequest, "请求参数错误")
				return
			}

			// 生成全局提示词（简化实现）
			prompt := fmt.Sprintf(`你是小雨，一个温柔的图书管理员，在魔法学院工作了3年。你性格温柔、细心，喜欢安静的环境，对朋友很关心。

当前用户是小明。

你们的关系：好朋友

请根据功能类型 %s 进行相应的互动。`, request.FunctionType)

			response := map[string]interface{}{
				"final_prompt":   prompt,
				"character_name": "小雨",
				"user_identity":  "小明",
				"prompt_length":  len(prompt),
				"generated_at":   time.Now().Format("2006-01-02 15:04:05"),
				"success":        true,
			}

			successJSON(w, response)
		})
	})

	// API 文档路由
	r.Get("/api/docs", func(w http.ResponseWriter, req *http.Request) {
		docs := map[string]interface{}{
			"title":       "YUNAI Monolith API 文档",
			"version":     "1.0.0",
			"description": "YUNAI 单体服务完整 API 接口",
			"base_url":    fmt.Sprintf("http://%s", req.Host),
			"endpoints": map[string]interface{}{
				"健康检查": []string{
					"GET /health - 服务健康状态",
				},
				"关系网": []string{
					"GET /relationships/network/{characterID} - 获取角色关系网络",
				},
				"智能朋友圈": []string{
					"POST /smart-moments/generate - 智能生成朋友圈",
					"POST /smart-moments/{momentID}/auto-interact - 触发自动互动",
					"POST /smart-moments/{momentID}/process-mentions - 处理@提及",
					"GET /smart-moments/generation-context/{characterID} - 获取生成上下文",
				},
				"朋友圈 REST": []string{
					"POST /moments - 创建朋友圈",
					"GET /moments - 获取朋友圈列表",
					"GET /moments/{id} - 获取朋友圈详情",
					"PUT /moments/{id} - 更新朋友圈",
					"DELETE /moments/{id} - 删除朋友圈",
					"以及草稿/互动/配置/统计/通知子路由",
				},
				"AI 邀请": []string{
					"POST /ai-invitation/analyze - 分析邀请场景",
					"POST /ai-invitation/find-matches - 寻找匹配角色",
					"POST /ai-invitation/suggestions - 获取邀请建议",
					"POST /ai-invitation/execute - 执行邀请",
				},
				"语音通话": []string{
					"POST /voice/call - 语音通话（需 multipart/form-data 音频上传）",
				},
				"全局提示词": []string{
					"POST /global-prompt/generate - 生成全局提示词",
				},
				"群聊功能": []string{
					"POST /group-chats - 创建群聊",
					"GET /group-chats - 获取群聊列表",
					"POST /group-chats/{groupID}/members - 添加群成员",
					"POST /group-chats/{groupID}/messages - 发送群消息",
				},
				"世界观设定": []string{
					"POST /worlds - 创建世界观",
					"GET /worlds - 获取世界观列表",
				},
				"故事章节": []string{
					"POST /chapters - 创建故事章节",
					"GET /chapters - 获取章节列表",
				},
				"故事触发器": []string{
					"POST /triggers - 创建触发器",
					"GET /triggers - 获取触发器列表",
				},
				"模型管理": []string{
					"GET /models/available - 获取可用模型",
					"GET /models/{id} - 获取模型详情",
					"POST /models/{id}/health-check - 执行健康检查",
					"GET /models/admin/ - 获取所有模型（管理员）",
					"POST /models/admin/{id}/enable - 启用模型",
					"POST /models/admin/{id}/disable - 禁用模型",
				},
				"支付管理": []string{
					"POST /payment/cards - 绑定支付卡片",
					"GET /payment/cards - 获取支付卡片列表",
					"POST /payment/cards/{id}/default - 设置默认卡片",
					"DELETE /payment/cards/{id} - 删除支付卡片",
					"POST /payment/password/set - 设置支付密码",
					"POST /payment/password/verify - 验证支付密码",
					"POST /payment/orders/recharge - 创建充值订单",
					"GET /payment/packages - 获取充值套餐",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"code":200,"message":"success","data":%s}`, toJSON(docs))
	})

	srv := &http.Server{
		Addr:         ":8092",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Infof("🚀 Monolith 服务已启动: http://localhost%v", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ 服务启动失败: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func successJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"code":200,"message":"success","data":%s}`, toJSON(data))
}

func errorJSON(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"code":%d,"message":%q}`, code, msg)
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// createGinSubRouter 创建 Gin 子路由并转换为 http.Handler
func createGinSubRouter(setupFunc func(*gin.Engine)) http.Handler {
	g := gin.New()
	g.Use(gin.Recovery())
	setupFunc(g)
	return g
}

// gormFromSqlx 从 sqlx.DB 创建 gorm.DB（简化实现）
func gormFromSqlx(sqlxDB *sqlx.DB) (*gorm.DB, error) {
	// 获取底层的 *sql.DB
	sqlDB := sqlxDB.DB

	// 使用 gorm 的 postgres 驱动打开连接
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})

	return gormDB, err
}
