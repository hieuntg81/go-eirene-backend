package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"bamboo-rescue/internal/config"
	"bamboo-rescue/internal/domain/entity"
	"bamboo-rescue/internal/domain/enum"
	"bamboo-rescue/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Sample data for generating realistic cases
var (
	// Ho Chi Minh City area coordinates
	minLat = 10.65
	maxLat = 10.90
	minLng = 106.55
	maxLng = 106.80

	// District names in HCMC
	districts = []string{
		"Quận 1", "Quận 2", "Quận 3", "Quận 4", "Quận 5", "Quận 6", "Quận 7", "Quận 8", "Quận 9", "Quận 10",
		"Quận 11", "Quận 12", "Quận Bình Tân", "Quận Bình Thạnh", "Quận Gò Vấp", "Quận Phú Nhuận",
		"Quận Tân Bình", "Quận Tân Phú", "Quận Thủ Đức", "Huyện Bình Chánh", "Huyện Củ Chi", "Huyện Hóc Môn",
	}

	// Street names
	streets = []string{
		"Nguyễn Huệ", "Lê Lợi", "Đồng Khởi", "Hai Bà Trưng", "Nguyễn Trãi", "Lê Văn Sỹ", "Nguyễn Văn Trỗi",
		"Cách Mạng Tháng 8", "Điện Biên Phủ", "Võ Văn Tần", "Trường Sa", "Hoàng Sa", "Phan Xích Long",
		"Nguyễn Đình Chiểu", "Lý Tự Trọng", "Pasteur", "Nam Kỳ Khởi Nghĩa", "Tôn Đức Thắng", "Võ Thị Sáu",
		"Nguyễn Thị Minh Khai", "Trần Hưng Đạo", "Nguyễn Công Trứ", "Bùi Viện", "Phạm Ngũ Lão",
	}

	// Animal case titles
	animalTitles = []string{
		"Chó bị thương cần giúp đỡ",
		"Mèo bị kẹt trên cây cao",
		"Chó con bị bỏ rơi",
		"Mèo hoang bị bệnh",
		"Chó bị xe đâm",
		"Mèo con cần cứu hộ",
		"Đàn mèo hoang cần thức ăn",
		"Chó già bị bỏ rơi",
		"Mèo bị mắc kẹt trong ống cống",
		"Chó bị lạc cần tìm chủ",
		"Mèo bị thương ở chân",
		"Chó con bị bỏ đói",
		"Đàn chó hoang hung dữ",
		"Mèo bị bỏ rơi trong hộp",
		"Chó bị xích không có nước",
	}

	// Flood case titles
	floodTitles = []string{
		"Ngập nước nghiêm trọng cần hỗ trợ",
		"Gia đình bị mắc kẹt do ngập",
		"Nước ngập vào nhà dân",
		"Cần di tản người già do ngập",
		"Ngập sâu không thể di chuyển",
		"Hộ dân cần thuyền cứu hộ",
		"Ngập úng khu dân cư",
		"Nước ngập đường không thể qua",
		"Cần hỗ trợ di tản trẻ em",
		"Nước tràn vào nhà tầng trệt",
		"Ngập nặng cần thực phẩm",
		"Gia đình mất điện do ngập",
		"Cần bơm nước ra khỏi nhà",
		"Ngập lụt kéo dài nhiều ngày",
		"Cần thuốc men cho người bệnh trong vùng ngập",
	}

	// Accident case titles
	accidentTitles = []string{
		"Tai nạn giao thông cần cấp cứu",
		"Người ngã từ độ cao",
		"Cháy nhà cần hỗ trợ",
		"Người bị kẹt trong xe",
		"Tai nạn xe máy nghiêm trọng",
		"Đuối nước cần cứu hộ",
		"Điện giật cần sơ cứu",
		"Va chạm xe ô tô",
		"Người bị ngất trên đường",
		"Tai nạn lao động",
		"Cháy cửa hàng",
		"Sập giàn giáo",
		"Xe tải đâm vào nhà",
		"Người bị thương do vật rơi",
		"Tai nạn tại công trường",
	}

	urgencyLevels = []enum.UrgencyLevel{
		enum.UrgencyLow,
		enum.UrgencyMedium,
		enum.UrgencyHigh,
		enum.UrgencyCritical,
	}

	caseStatuses = []enum.CaseStatus{
		enum.CaseStatusPending,
		enum.CaseStatusAccepted,
		enum.CaseStatusInProgress,
		enum.CaseStatusResolved,
	}

	animalTypes = []enum.AnimalType{
		enum.AnimalTypeDog,
		enum.AnimalTypeCat,
		enum.AnimalTypeBird,
		enum.AnimalTypeOther,
	}

	animalConditions = []enum.AnimalCondition{
		enum.AnimalConditionInjured,
		enum.AnimalConditionTrapped,
		enum.AnimalConditionSick,
		enum.AnimalConditionAbandoned,
	}

	accidentTypes = []enum.AccidentType{
		enum.AccidentTypeTraffic,
		enum.AccidentTypeFall,
		enum.AccidentTypeFire,
		enum.AccidentTypeDrowning,
		enum.AccidentTypeElectric,
	}

	reporterNames = []string{
		"Nguyễn Văn An", "Trần Thị Bình", "Lê Văn Cường", "Phạm Thị Dung", "Hoàng Văn Em",
		"Vũ Thị Phương", "Đặng Văn Giang", "Bùi Thị Hương", "Ngô Văn Khoa", "Đỗ Thị Lan",
		"Trịnh Văn Minh", "Đinh Thị Ngọc", "Lý Văn Phong", "Mai Thị Quỳnh", "Hồ Văn Sơn",
	}
)

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Initialize logger
	log, _ := zap.NewProduction()
	defer log.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database
	db, err := database.NewPostgresDB(&cfg.Database, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close(db)

	log.Info("Starting to seed 1000 cases...")

	// Seed cases
	if err := seedCases(db, 1000, log); err != nil {
		log.Fatal("Failed to seed cases", zap.Error(err))
	}

	log.Info("Successfully seeded 1000 cases!")
}

func seedCases(db *gorm.DB, count int, log *zap.Logger) error {
	ctx := context.Background()
	batchSize := 100

	for i := 0; i < count; i += batchSize {
		end := i + batchSize
		if end > count {
			end = count
		}

		cases := make([]*entity.Case, 0, end-i)
		for j := i; j < end; j++ {
			c := generateRandomCase(j)
			cases = append(cases, c)
		}

		// Insert cases in batch
		for _, c := range cases {
			if err := createCaseWithDetails(ctx, db, c); err != nil {
				log.Warn("Failed to create case", zap.Error(err), zap.Int("index", i))
				continue
			}
		}

		log.Info("Progress", zap.Int("created", end), zap.Int("total", count))
	}

	return nil
}

func createCaseWithDetails(ctx context.Context, db *gorm.DB, c *entity.Case) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the main case
		if err := tx.Omit("AnimalDetails", "FloodDetails", "AccidentDetails", "Media", "Volunteers", "Updates").Create(c).Error; err != nil {
			return err
		}

		// Create type-specific details
		switch c.CaseType {
		case enum.CaseTypeAnimal:
			if c.AnimalDetails != nil {
				c.AnimalDetails.CaseID = c.ID
				if err := tx.Create(c.AnimalDetails).Error; err != nil {
					return err
				}
			}
		case enum.CaseTypeFlood:
			if c.FloodDetails != nil {
				c.FloodDetails.CaseID = c.ID
				if err := tx.Create(c.FloodDetails).Error; err != nil {
					return err
				}
			}
		case enum.CaseTypeAccident:
			if c.AccidentDetails != nil {
				c.AccidentDetails.CaseID = c.ID
				if err := tx.Create(c.AccidentDetails).Error; err != nil {
					return err
				}
			}
		}

		// Create initial update
		update := &entity.CaseUpdate{
			ID:         uuid.New(),
			CaseID:     c.ID,
			UpdateType: enum.UpdateTypeSystem,
			Content:    stringPtr("Case created"),
			NewStatus:  &c.Status,
		}
		return tx.Create(update).Error
	})
}

func generateRandomCase(index int) *entity.Case {
	caseTypes := []enum.CaseType{enum.CaseTypeAnimal, enum.CaseTypeFlood, enum.CaseTypeAccident}
	caseType := caseTypes[rand.Intn(len(caseTypes))]

	// Generate random location in HCMC
	lat := minLat + rand.Float64()*(maxLat-minLat)
	lng := minLng + rand.Float64()*(maxLng-minLng)

	// Generate address
	streetNum := rand.Intn(500) + 1
	street := streets[rand.Intn(len(streets))]
	district := districts[rand.Intn(len(districts))]
	address := fmt.Sprintf("%d %s, %s, TP. Hồ Chí Minh", streetNum, street, district)

	// Generate title based on case type
	var title string
	switch caseType {
	case enum.CaseTypeAnimal:
		title = animalTitles[rand.Intn(len(animalTitles))]
	case enum.CaseTypeFlood:
		title = floodTitles[rand.Intn(len(floodTitles))]
	case enum.CaseTypeAccident:
		title = accidentTitles[rand.Intn(len(accidentTitles))]
	}

	// Add some variation to title
	title = fmt.Sprintf("%s - Case #%d", title, index+1)

	// Generate urgency with weighted distribution
	urgency := getWeightedUrgency()

	// Generate status with weighted distribution
	status := getWeightedStatus()

	// Generate phone number
	phone := fmt.Sprintf("09%d%d%d%d%d%d%d%d",
		rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10),
		rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10))

	// Generate reporter name
	reporterName := reporterNames[rand.Intn(len(reporterNames))]

	// Generate random creation time within the last 30 days
	daysAgo := rand.Intn(30)
	hoursAgo := rand.Intn(24)
	createdAt := time.Now().AddDate(0, 0, -daysAgo).Add(-time.Duration(hoursAgo) * time.Hour)

	c := &entity.Case{
		ID:             uuid.New(),
		CaseType:       caseType,
		Status:         status,
		Urgency:        urgency,
		Latitude:       lat,
		Longitude:      lng,
		Address:        &address,
		Title:          title,
		Description:    stringPtr(fmt.Sprintf("Mô tả chi tiết cho case %s. Cần hỗ trợ khẩn cấp tại địa chỉ trên.", title)),
		ReporterName:   &reporterName,
		ReporterPhone:  phone,
		IsAnonymous:    rand.Float32() < 0.1, // 10% anonymous
		VolunteerCount: 0,
		MaxVolunteers:  5,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}

	// Set volunteer count based on status
	if status == enum.CaseStatusAccepted || status == enum.CaseStatusInProgress {
		c.VolunteerCount = rand.Intn(3) + 1
		c.AcceptedAt = timePtr(createdAt.Add(time.Duration(rand.Intn(60)) * time.Minute))
	} else if status == enum.CaseStatusResolved {
		c.VolunteerCount = rand.Intn(5) + 1
		c.AcceptedAt = timePtr(createdAt.Add(time.Duration(rand.Intn(60)) * time.Minute))
		c.ResolvedAt = timePtr(createdAt.Add(time.Duration(rand.Intn(24)) * time.Hour))
	}

	// Add type-specific details
	switch caseType {
	case enum.CaseTypeAnimal:
		c.AnimalDetails = &entity.CaseAnimalDetails{
			ID:             uuid.New(),
			AnimalType:     animalTypes[rand.Intn(len(animalTypes))],
			Condition:      animalConditions[rand.Intn(len(animalConditions))],
			EstimatedCount: rand.Intn(5) + 1,
		}
	case enum.CaseTypeFlood:
		peopleCount := rand.Intn(10) + 1
		waterLevel := rand.Intn(200) + 10
		c.FloodDetails = &entity.CaseFloodDetails{
			ID:           uuid.New(),
			PeopleCount:  &peopleCount,
			HasChildren:  rand.Float32() < 0.3,
			HasElderly:   rand.Float32() < 0.4,
			HasDisabled:  rand.Float32() < 0.1,
			WaterLevelCm: &waterLevel,
		}
	case enum.CaseTypeAccident:
		c.AccidentDetails = &entity.CaseAccidentDetails{
			ID:             uuid.New(),
			AccidentType:   accidentTypes[rand.Intn(len(accidentTypes))],
			VictimCount:    rand.Intn(5) + 1,
			HasUnconscious: rand.Float32() < 0.2,
			HasBleeding:    rand.Float32() < 0.4,
			HasFracture:    rand.Float32() < 0.3,
			IsTrapped:      rand.Float32() < 0.15,
		}
	}

	return c
}

func getWeightedUrgency() enum.UrgencyLevel {
	r := rand.Float32()
	if r < 0.1 {
		return enum.UrgencyCritical // 10%
	} else if r < 0.35 {
		return enum.UrgencyHigh // 25%
	} else if r < 0.75 {
		return enum.UrgencyMedium // 40%
	}
	return enum.UrgencyLow // 25%
}

func getWeightedStatus() enum.CaseStatus {
	r := rand.Float32()
	if r < 0.4 {
		return enum.CaseStatusPending // 40%
	} else if r < 0.6 {
		return enum.CaseStatusAccepted // 20%
	} else if r < 0.8 {
		return enum.CaseStatusInProgress // 20%
	}
	return enum.CaseStatusResolved // 20%
}

func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
