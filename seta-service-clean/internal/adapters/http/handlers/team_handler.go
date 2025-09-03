package handlers

import (
	"net/http"
	"seta/internal/adapters/http/utils"
	"seta/internal/application"
	"seta/internal/domain/common"
	"seta/internal/domain/team"
	"seta/pkg/errorHandling"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TeamHandler struct {
	createTeamUseCase    *application.CreateTeamUseCase
	addMemberUseCase     *application.AddMemberUseCase
	removeMemberUseCase  *application.RemoveMemberUseCase
	addManagerUseCase    *application.AddManagerUseCase
	removeManagerUseCase *application.RemoveManagerUseCase
	getTeamAssetsUseCase *application.GetTeamAssetsUseCase
}

func NewTeamHandler(
	createUC *application.CreateTeamUseCase, addMemberUC *application.AddMemberUseCase,
	removeMemberUC *application.RemoveMemberUseCase, addManagerUC *application.AddManagerUseCase,
	removeManagerUC *application.RemoveManagerUseCase, getTeamAssetsUC *application.GetTeamAssetsUseCase,
) *TeamHandler {
	return &TeamHandler{
		createTeamUseCase: createUC, addMemberUseCase: addMemberUC,
		removeMemberUseCase: removeMemberUC, addManagerUseCase: addManagerUC,
		removeManagerUseCase: removeManagerUC, getTeamAssetsUseCase: getTeamAssetsUC,
	}
}

// DTOs
type createTeamRequest struct {
	TeamName string `json:"teamName" binding:"required"`
	Managers []struct {
		ManagerID uuid.UUID `json:"managerId" binding:"required"`
		IsLead    bool      `json:"isLead"`
	} `json:"managers" binding:"required,min=1"`
	Members []struct {
		MemberID uuid.UUID `json:"memberId" binding:"required"`
	} `json:"members"`
}

type addRemoveUserInput struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
}

// Handlers
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req createTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()}); return
	}
	creatorID, ok := utils.GetUserIDFromContext(c); if !ok { return }

	managers := make([]team.Manager, len(req.Managers))
	for i, m := range req.Managers {
		managers[i] = team.Manager{UserID: common.UserID(m.ManagerID), IsLead: m.IsLead}
	}
	members := make([]team.Member, len(req.Members))
	for i, m := range req.Members {
		members[i] = team.Member{UserID: common.UserID(m.MemberID)}
	}

	input := application.CreateTeamInput{
		TeamName: req.TeamName, CreatorID: creatorID, Managers: managers, Members: members,
	}
	newTeam, err := h.createTeamUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()}); return
	}
	c.JSON(http.StatusCreated, gin.H{"teamId": newTeam.ID, "teamName": newTeam.Name})
}

func (h *TeamHandler) AddMember(c *gin.Context) {
	teamID, ok := utils.GetUUIDFromParam(c, "teamId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }

	var req addRemoveUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()}); return
	}

	input := application.AddMemberInput{
		TeamID: common.TeamID(teamID), MemberID: common.UserID(req.UserID), RequesterID: requesterID,
	}
	if err := h.addMemberUseCase.Execute(c.Request.Context(), input); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.Status(http.StatusNoContent)
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID, ok := utils.GetUUIDFromParam(c, "teamId"); if !ok { return }
	memberID, ok := utils.GetUUIDFromParam(c, "memberId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }

	input := application.RemoveMemberInput{
		TeamID: common.TeamID(teamID), MemberID: common.UserID(memberID), RequesterID: requesterID,
	}
	if err := h.removeMemberUseCase.Execute(c.Request.Context(), input); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.Status(http.StatusNoContent)
}

func (h *TeamHandler) AddManager(c *gin.Context) {
	teamID, ok := utils.GetUUIDFromParam(c, "teamId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }

	var req addRemoveUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()}); return
	}

	input := application.AddManagerInput{
		TeamID: common.TeamID(teamID), ManagerID: common.UserID(req.UserID), RequesterID: requesterID,
	}
	if err := h.addManagerUseCase.Execute(c.Request.Context(), input); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.Status(http.StatusNoContent)
}

func (h *TeamHandler) RemoveManager(c *gin.Context) {
	teamID, ok := utils.GetUUIDFromParam(c, "teamId"); if !ok { return }
	managerID, ok := utils.GetUUIDFromParam(c, "managerId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }

	input := application.RemoveManagerInput{
		TeamID: common.TeamID(teamID), ManagerID: common.UserID(managerID), RequesterID: requesterID,
	}
	if err := h.removeManagerUseCase.Execute(c.Request.Context(), input); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.Status(http.StatusNoContent)
}

func (h *TeamHandler) GetTeamAssets(c *gin.Context) {
	teamID, ok := utils.GetUUIDFromParam(c, "teamId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }

	input := application.GetTeamAssetsInput{
		TeamID: common.TeamID(teamID), RequesterID: requesterID,
	}
	assets, err := h.getTeamAssetsUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.JSON(http.StatusOK, assets)
}
