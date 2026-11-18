package services

import (
	"errors"
	"strings"

	"Afaisal94/first-app/internal/model"
)

// CategoryService mendefinisikan kontrak business logic (usecase) untuk Category.
// Handler hanya bergantung pada interface ini.
// Service hanya mengimpor model — tidak tahu gorm, gin, maupun dto
// (dependency rule: usecase tidak bergantung ke lapisan luar).
type CategoryService interface {
	FindAll() ([]model.Category, error)
	FindByID(id string) (*model.Category, error)
	Create(category string) (*model.Category, error)
	Update(id, category string) (*model.Category, error)
	Delete(id string) error
}

type categoryService struct {
	repo model.CategoryRepository
}

// NewCategoryService membuat implementasi CategoryService dengan port yang di-inject.
func NewCategoryService(repo model.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) FindAll() ([]model.Category, error) {
	return s.repo.FindAll()
}

func (s *categoryService) FindByID(id string) (*model.Category, error) {
	return s.repo.FindByID(id)
}

func (s *categoryService) Create(category string) (*model.Category, error) {
	category = strings.TrimSpace(category)
	if category == "" {
		return nil, errors.New("category must not be empty")
	}

	c := &model.Category{
		Category: category,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *categoryService) Update(id, category string) (*model.Category, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	category = strings.TrimSpace(category)
	if category == "" {
		return nil, errors.New("category must not be empty")
	}

	c.Category = category
	if err := s.repo.Update(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *categoryService) Delete(id string) error {
	// Validasi dulu bahwa category ada.
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
