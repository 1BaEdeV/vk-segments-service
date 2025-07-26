package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Database connection settings
const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "your_user"
	dbPassword = "your_password"
	dbName     = "segments"
)

// SegmentService содержит методы для работы с сегментами
type SegmentService struct {
	db *sql.DB
}

// NewSegmentService создает новый экземпляр SegmentService
func NewSegmentService(db *sql.DB) *SegmentService {
	return &SegmentService{db: db}
}

// CreateSegment создает новый сегмент
func (s *SegmentService) CreateSegment(slug string) error {
	_, err := s.db.Exec("INSERT INTO segments (slug) VALUES ($1) ON CONFLICT (slug) DO NOTHING", slug)
	if err != nil {
		return fmt.Errorf("failed to create segment: %w", err)
	}
	return nil
}

// DeleteSegment удаляет сегмент
func (s *SegmentService) DeleteSegment(slug string) error {
	_, err := s.db.Exec("DELETE FROM segments WHERE slug = $1", slug)
	if err != nil {
		return fmt.Errorf("failed to delete segment: %w", err)
	}
	return nil
}

// AddUserToSegment добавляет пользователя в сегмент
func (s *SegmentService) AddUserToSegment(userID int, segmentSlug string) error {
	// Проверяем существование пользователя
	var userExists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&userExists)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if !userExists {
		_, err = s.db.Exec("INSERT INTO users (id) VALUES ($1)", userID)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	_, err = s.db.Exec(`
		INSERT INTO user_segments (user_id, segment_slug) 
		VALUES ($1, $2) 
		ON CONFLICT (user_id, segment_slug) DO NOTHING`,
		userID, segmentSlug)
	if err != nil {
		return fmt.Errorf("failed to add user to segment: %w", err)
	}

	return nil
}

// RemoveUserFromSegment удаляет пользователя из сегмента
func (s *SegmentService) RemoveUserFromSegment(userID int, segmentSlug string) error {
	_, err := s.db.Exec(`
		DELETE FROM user_segments 
		WHERE user_id = $1 AND segment_slug = $2`,
		userID, segmentSlug)
	if err != nil {
		return fmt.Errorf("failed to remove user from segment: %w", err)
	}
	return nil
}

// GetUserSegments возвращает сегменты пользователя
func (s *SegmentService) GetUserSegments(userID int) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT segment_slug 
		FROM user_segments 
		WHERE user_id = $1`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user segments: %w", err)
	}
	defer rows.Close()

	var segments []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, fmt.Errorf("failed to scan segment: %w", err)
		}
		segments = append(segments, slug)
	}

	return segments, nil
}

// DistributeSegment распределяет сегмент среди % пользователей
func (s *SegmentService) DistributeSegment(segmentSlug string, percent int) error {
	// Проверяем существование сегмента
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM segments WHERE slug = $1)", segmentSlug).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check segment existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("segment %s does not exist", segmentSlug)

		// Остальная логика метода...
	}
	// Получаем общее количество пользователей
	var totalUsers int
	err = s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if totalUsers == 0 {
		return fmt.Errorf("no users available")
	}

	// Вычисляем количество пользователей для добавления
	usersToAdd := (totalUsers * percent) / 100
	if usersToAdd < 1 {
		usersToAdd = 1
	}

	// Выбираем случайных пользователей
	rows, err := s.db.Query(`
		SELECT id 
		FROM users 
		ORDER BY RANDOM() 
		LIMIT $1`,
		usersToAdd)
	if err != nil {
		return fmt.Errorf("failed to select random users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return fmt.Errorf("failed to scan user ID: %w", err)
		}

		if err := s.AddUserToSegment(userID, segmentSlug); err != nil {
			return fmt.Errorf("failed to add segment to user %d: %w", userID, err)
		}
	}

	return nil
}

func main() {
	// Подключение к БД
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Проверка подключения
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL")

	// Инициализация сервиса
	service := NewSegmentService(db)

	// Инициализация Gin роутера
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	r.GET("/segments", func(c *gin.Context) {
		rows, err := db.Query("SELECT slug FROM segments")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var segments []string
		for rows.Next() {
			var slug string
			if err := rows.Scan(&slug); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			segments = append(segments, slug)
		}

		c.JSON(http.StatusOK, gin.H{"segments": segments})
	})

	// Endpoint для создания сегмента
	r.POST("/segments", func(c *gin.Context) {
		var input struct {
			Slug string `json:"slug" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := service.CreateSegment(input.Slug); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"status": "segment created"})
	})

	// Endpoint для удаления сегмента
	r.DELETE("/segments/:slug", func(c *gin.Context) {
		slug := c.Param("slug")

		if err := service.DeleteSegment(slug); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "segment deleted"})
	})

	// Endpoint для добавления пользователя в сегмент
	r.POST("/users/:id/segments", func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		var input struct {
			SegmentSlug string `json:"segment_slug" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := service.AddUserToSegment(userID, input.SegmentSlug); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "user added to segment"})
	})

	// Endpoint для удаления пользователя из сегмента
	r.DELETE("/users/:id/segments/:slug", func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		slug := c.Param("slug")

		if err := service.RemoveUserFromSegment(userID, slug); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "user removed from segment"})
	})

	// Endpoint для получения сегментов пользователя
	r.GET("/users/:id/segments", func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		segments, err := service.GetUserSegments(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"segments": segments})
	})

	// Endpoint для распределения сегмента
	r.POST("/segments/:slug/distribute", func(c *gin.Context) {
		slug := c.Param("slug")

		var input struct {
			Percent int `json:"percent" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := service.DistributeSegment(slug, input.Percent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "segment distributed"})
	})

	// Запуск сервера
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
