package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func RouterSetup(timeout time.Duration, db *mongo.Database, router *gin.Engine) {
	publicRouter := router.Group("")
	// All public APIS
	NewPublicRouter(db, timeout, publicRouter)

	roleRouter := router.Group("")
	NewRoleRouter(db, timeout, roleRouter)

	profileRouter := router.Group("")
	NewProfileRouter(db, timeout, profileRouter)

	countryRouter := router.Group("")
	NewCountryRouter(db, timeout, countryRouter)

	travelRouter := router.Group("")
	NewTravelPurposeRouter(db, timeout, travelRouter)

	processRouter := router.Group("")
	NewProcessRouter(db, timeout, processRouter)

	subprocessRouter := router.Group("")
	NewSubprocessRouter(db, timeout, subprocessRouter)

	departmentRouter := router.Group("")
	NewDepartmentRouter(db, timeout, departmentRouter)

	currencyRouter := router.Group("")
	NewCurrencyRouter(db, timeout, currencyRouter)

	requestRouter := router.Group("")
	NewRequestRouter(db, timeout, requestRouter)

	districtRouter := router.Group("")
	NewDistrictRouter(db, timeout, districtRouter)

	branchRouter := router.Group("")
	NewBranchRouter(db, timeout, branchRouter)
}
