package repository

import (
	"errors"
	"regexp"
	"strconv"
	"sync"

	"github.com/Kev2406/PEA/internal/domain/model"
	"gorm.io/gorm"
)

type SealRepository struct {
	db *gorm.DB
}

func NewSealRepository(db *gorm.DB) *SealRepository {
	return &SealRepository{db: db}
}

func (r *SealRepository) Create(seal *model.Seal) error {
	return r.db.Create(seal).Error
}

func (r *SealRepository) CreateMultiple(seals []model.Seal) error {
	// If no seals to insert, return early
	if len(seals) == 0 {
		return nil
	}

	// Extract all seal numbers to check
	var sealNumbers []string
	for _, seal := range seals {
		sealNumbers = append(sealNumbers, seal.SealNumber)
	}

	// Find existing seals
	var existingSeals []model.Seal
	if err := r.db.Where("seal_number IN ?", sealNumbers).Find(&existingSeals).Error; err != nil {
		return err
	}

	// Create a map for faster lookup
	existingSealMap := make(map[string]bool)
	for _, seal := range existingSeals {
		existingSealMap[seal.SealNumber] = true
	}

	// Filter out existing seals
	var newSeals []model.Seal
	for _, seal := range seals {
		if !existingSealMap[seal.SealNumber] {
			newSeals = append(newSeals, seal)
		}
	}

	// If no new seals to insert, return success
	if len(newSeals) == 0 {
		return nil
	}

	// Insert only the new seals
	return r.db.Create(&newSeals).Error
}

func (r *SealRepository) FindByNumber(sealNumber string) (*model.Seal, error) {
	var seal model.Seal
	if err := r.db.Where("seal_number = ?", sealNumber).First(&seal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("seal not found")
		}
		return nil, err
	}
	return &seal, nil
}

func (r *SealRepository) GetLatestSeal() (*model.Seal, error) {
	var seals []model.Seal
	if err := r.db.Order("seal_number DESC").Find(&seals).Error; err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`^([A-Za-z]*)(\d+)$`)

	var latestSeal *model.Seal
	var maxNumber int64
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := range seals {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			matches := re.FindStringSubmatch(seals[i].SealNumber)
			if len(matches) == 3 {
				num := parseInt64(matches[2])
				mu.Lock()
				if num > maxNumber {
					maxNumber = num
					latestSeal = &seals[i]
				}
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if latestSeal == nil {
		return nil, nil
	}
	return latestSeal, nil
}

func (r *SealRepository) FindByPrefix(prefix string) (*model.Seal, error) {
	var seal model.Seal
	if err := r.db.Where("seal_number LIKE ?", prefix+"%").
		Order("LENGTH(seal_number) DESC, seal_number DESC").
		First(&seal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &seal, nil
}

func (r *SealRepository) Update(seal *model.Seal) error {
	return r.db.Save(seal).Error
}

func (r *SealRepository) Delete(sealID uint) error {
	return r.db.Delete(&model.Seal{}, sealID).Error
}

func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
func (r *SealRepository) CheckSealExists(sealNumber string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Seal{}).Where("seal_number = ?", sealNumber).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// func (r *SealRepository) CheckSealExists(sealNumber string) (bool, error) {
//     var count int64
//     if err := r.db.Model(&model.Seal{}).
//        Where("seal_number = ?", sealNumber).
//        Count(&count).Error; err != nil {
//         return false, err
//     }
//     return count > 0, nil
// }
