package handler

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/vadimbarashkov/url-shortener/internal/storage"
)

type AddURLRequest struct {
	Alias string `json:"alias" validate:"required"`
	URL   string `json:"url" validate:"required,url"`
}

type UpdateURLRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type URLHandler struct {
	storage  storage.URLStorage
	validate *validator.Validate
}

func NewURLHandler(storage storage.URLStorage) *URLHandler {
	return &URLHandler{
		storage:  storage,
		validate: validator.New(),
	}
}

func (h *URLHandler) Add(c *fiber.Ctx) error {
	var req AddURLRequest

	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	// TODO: add detailed error response
	if err := h.validate.Struct(req); err != nil {
		return fiber.ErrBadRequest
	}

	err := h.storage.Add(c.Context(), req.Alias, req.URL)
	if err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			return fiber.ErrConflict
		}

		return fiber.ErrInternalServerError
	}

	return c.SendStatus(fiber.StatusCreated)
}

func (h *URLHandler) Get(c *fiber.Ctx) error {
	alias := c.Params("alias")

	url, err := h.storage.Get(c.Context(), alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			return fiber.ErrNotFound
		}

		return fiber.ErrInternalServerError
	}

	return c.Redirect(url, fiber.StatusFound)
}

func (h *URLHandler) Update(c *fiber.Ctx) error {
	alias := c.Params("alias")

	var req UpdateURLRequest

	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	// TODO: add detailed error response
	if err := h.validate.Struct(req); err != nil {
		return fiber.ErrBadRequest
	}

	err := h.storage.Update(c.Context(), alias, req.URL)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			return fiber.ErrNotFound
		}

		return fiber.ErrInternalServerError
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *URLHandler) Delete(c *fiber.Ctx) error {
	alias := c.Params("alias")

	err := h.storage.Delete(c.Context(), alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			return fiber.ErrNotFound
		}

		return fiber.ErrInternalServerError
	}

	return c.SendStatus(fiber.StatusNoContent)
}
