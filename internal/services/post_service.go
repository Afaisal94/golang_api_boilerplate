package services

import (
	"errors"
	"strconv"
	"strings"

	"Afaisal94/first-app/internal/model"
)

// PostService mendefinisikan kontrak business logic (usecase) untuk Post.
// Handler hanya bergantung pada interface ini.
// Service hanya mengimpor model — tidak tahu gorm, gin, maupun dto
// (dependency rule: usecase tidak bergantung ke lapisan luar).
type PostService interface {
	FindAll() ([]model.Post, error)
	FindByID(id string) (*model.Post, error)
	Create(title, content string, categoryID uint) (*model.Post, error)
	Update(id, title, content string, categoryID uint) (*model.Post, error)
	Delete(id string) error
}

type postService struct {
	repo         model.PostRepository
	categoryRepo model.CategoryRepository
}

// NewPostService membuat implementasi PostService dengan port yang di-inject.
// categoryRepo dibutuhkan untuk memvalidasi category_id sebelum write.
func NewPostService(repo model.PostRepository, categoryRepo model.CategoryRepository) PostService {
	return &postService{repo: repo, categoryRepo: categoryRepo}
}

func (s *postService) FindAll() ([]model.Post, error) {
	return s.repo.FindAll()
}

func (s *postService) FindByID(id string) (*model.Post, error) {
	return s.repo.FindByID(id)
}

func (s *postService) Create(title, content string, categoryID uint) (*model.Post, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" || content == "" {
		return nil, errors.New("title and content must not be empty")
	}
	if categoryID == 0 {
		return nil, errors.New("category_id must not be empty")
	}

	// Validasi relasi: pastikan category tujuan ada (termasuk belum soft-deleted).
	category, err := s.categoryRepo.FindByID(strconv.FormatUint(uint64(categoryID), 10))
	if err != nil {
		return nil, err
	}

	post := &model.Post{
		Title:      title,
		Content:    content,
		CategoryID: categoryID,
		Category:   *category,
	}
	if err := s.repo.Create(post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *postService) Update(id, title, content string, categoryID uint) (*model.Post, error) {
	post, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" || content == "" {
		return nil, errors.New("title and content must not be empty")
	}
	if categoryID == 0 {
		return nil, errors.New("category_id must not be empty")
	}

	// Validasi relasi: pastikan category tujuan ada (termasuk belum soft-deleted).
	category, err := s.categoryRepo.FindByID(strconv.FormatUint(uint64(categoryID), 10))
	if err != nil {
		return nil, err
	}

	post.Title = title
	post.Content = content
	post.CategoryID = categoryID
	post.Category = *category
	if err := s.repo.Update(post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *postService) Delete(id string) error {
	// Validasi dulu bahwa post ada.
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
