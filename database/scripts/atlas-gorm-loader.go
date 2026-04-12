package main

import (
	"fmt"
	"io"
	"os"

	_ "ariga.io/atlas-go-sdk/recordriver"
	"ariga.io/atlas-provider-gorm/gormschema"
	"encore.app/database/models"
)

// Define the models to generate migrations for.
var model_list = []any{
	&models.Business{},
	&models.Outlet{},
	&models.User{},
	&models.Customer{},
	&models.Membership{},
	&models.CustomerMembership{},
	&models.MembershipBenefit{},
	&models.MembershipUpgradeRule{},
	&models.CustomerMembershipStats{},
	&models.CustomerProductPurchaseStats{},
	&models.Role{},
	&models.PointRule{},
	&models.PointTransaction{},
	&models.GroupRole{},
	&models.Permission{},
	&models.PointRedemptionRule{},
	&models.PointRedemptionTransaction{},
	&models.Discount{},
	&models.Voucher{},
	&models.PermissionPreset{},
	&models.ProductCategory{},
	&models.ProductSubCategory{},
	&models.Product{},
	&models.ProductCategoryMapping{},
	&models.Ingredient{},
	&models.Stock{},
	&models.Recipe{},
	&models.RecipeStep{},
	&models.Order{},
	&models.OrderItem{},
	&models.Transaction{},
	&models.MerchantSecret{},
	&models.ActivityLog{},
	&models.ProductModifierMapping{},
	&models.ModifierGroups{},
	&models.ModifierOptions{},
	&models.ProductIngredientMapping{},
	&models.StockReport{},
	&models.SelectedModifierGroup{},
	&models.ExpensesOutlet{},
	&models.ModifierIngredientMapping{},
	&models.StockRequest{},
	&models.StockRequestedItem{},
	&models.SystemData{},
	&models.OutletProduct{},
	&models.Cache{},
	&models.OrderDetails{},
	&models.BusinessConfiguration{},
	&models.Notification{},
	&models.AppVersion{},
	&models.UserToken{},
	&models.OutletModifierOption{},
	&models.VoucherOutlet{},
	&models.CustomerToken{},
	&models.ProductWastageReport{},
	&models.ProductWastageType{},
	&models.CustomerVoucher{},
	&models.OutletGroup{},
	&models.CustomerFavouriteProduct{},
	&models.Mission{},
	&models.MissionCriteria{},
	&models.MissionReward{},
	&models.CustomerMission{},
	&models.CustomerMissionCriteriaProgress{},
	&models.MissionRewardGrant{},
	&models.Onboarding{},
	&models.Announcement{},
	&models.Delivery{},
	&models.FeedbackQuestion{},
	&models.CustomerDeliveryAddress{},
	&models.Feedback{},
	&models.PaymentMethodConfiguration{},
	&models.VoucherEligibilityRule{},
	&models.DigitalSignageContent{},
	&models.DigitalSignageSlide{},
	&models.DigitalSignageSlideItem{},
	&models.OutletOperationSchedule{},
	&models.OutletOperationTimeSlot{},
	&models.OutletTerminal{},
	&models.CampaignPushNotification{},
	&models.NotificationQueue{},
	//&models.Addon{},
	//&models.ProductAddonMapping{},ß
}

func main() {
	stmts, err := gormschema.New("postgres").Load(model_list...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
