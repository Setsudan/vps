package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"launay-dot-one/middlewares"
	m "launay-dot-one/models"
	resumesvc "launay-dot-one/services/resumes"
	"launay-dot-one/utils"
)

type ResumeController struct {
	svc    resumesvc.Service
	logger *logrus.Logger
}

func NewResumeController(svc resumesvc.Service, logger *logrus.Logger) *ResumeController {
	return &ResumeController{svc: svc, logger: logger}
}

// RegisterRoutes wires up:
//   - GET/POST/PUT/DELETE   /resume     (auth required)
//   - GET                     /resumes/:user_id  (public)
func (rc *ResumeController) RegisterRoutes(r *gin.Engine) {
	grp := r.Group("/resume", middlewares.AuthMiddleware())
	{
		grp.GET("", rc.GetMyResume)
		grp.POST("", rc.CreateResume)
		grp.PUT("", rc.UpdateResume)
		grp.DELETE("", rc.DeleteResume)
	}
	// public lookup
	r.GET("/resumes/:user_id", rc.GetResumeByUser)
}

// GetMyResume returns the authenticated user's resume.
func (rc *ResumeController) GetMyResume(c *gin.Context) {
	userID := c.GetString("user_id")
	res, err := rc.svc.GetByUser(c.Request.Context(), userID)
	if err != nil {
		rc.logger.Error("GetMyResume error: ", err)
		utils.RespondError(c, http.StatusNotFound, "Resume not found", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Resume fetched", res)
}

// CreateResume creates a resume for the current user.
func (rc *ResumeController) CreateResume(c *gin.Context) {
	var res m.Resume
	if err := c.ShouldBindJSON(&res); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	res.UserID = c.GetString("user_id")
	if err := rc.svc.Create(c.Request.Context(), &res); err != nil {
		rc.logger.Error("CreateResume error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to create resume", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusCreated, "Resume created", res)
}

// UpdateResume updates the authenticated user's resume.
func (rc *ResumeController) UpdateResume(c *gin.Context) {
	var res m.Resume
	if err := c.ShouldBindJSON(&res); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	// ensure user can only update their own resume
	res.UserID = c.GetString("user_id")
	if err := rc.svc.Update(c.Request.Context(), &res); err != nil {
		rc.logger.Error("UpdateResume error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to update resume", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Resume updated", res)
}

// DeleteResume removes the authenticated user's resume.
func (rc *ResumeController) DeleteResume(c *gin.Context) {
	userID := c.GetString("user_id")
	if err := rc.svc.DeleteByUser(c.Request.Context(), userID); err != nil {
		rc.logger.Error("DeleteResume error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to delete resume", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Resume deleted", nil)
}

// GetResumeByUser fetches a public view of another user's resume.
func (rc *ResumeController) GetResumeByUser(c *gin.Context) {
	uid := c.Param("user_id")
	res, err := rc.svc.GetByUser(c.Request.Context(), uid)
	if err != nil {
		rc.logger.Error("GetResumeByUser error: ", err)
		utils.RespondError(c, http.StatusNotFound, "Resume not found", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Resume fetched", res)
}
