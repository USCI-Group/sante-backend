package constants

import "encore.dev/types/uuid"

type ActionPermission string

// List of all action permissions
const (
	// USER MANAGEMENT ACTIONS
	CreateUserAction     ActionPermission = "create_user"
	ReadUserAction       ActionPermission = "read_user"
	UpdateUserAction     ActionPermission = "update_user"
	DeleteUserAction     ActionPermission = "delete_user"
	ResetPasswordAction  ActionPermission = "reset_password"
	UserPermissionAction ActionPermission = "user_permission"
	ActivityLogAction    ActionPermission = "activity_log"
	CreateRoleAction     ActionPermission = "create_role"
	ReadRoleAction       ActionPermission = "read_role"
	UpdateRoleAction     ActionPermission = "update_role"
	DeleteRoleAction     ActionPermission = "delete_role"

	// BUSINESS MANAGEMENT ACTIONS
	CreateBusinessAction ActionPermission = "create_business"
	ReadBusinessAction   ActionPermission = "read_business"
	UpdateBusinessAction ActionPermission = "update_business"
	DeleteBusinessAction ActionPermission = "delete_business"

	// OUTLET MANAGEMENT ACTIONS
	CreateOutletAction ActionPermission = "create_outlet"
	ReadOutletAction   ActionPermission = "read_outlet"
	UpdateOutletAction ActionPermission = "update_outlet"
	DeleteOutletAction ActionPermission = "delete_outlet"

	/* CRUDOrderAction ActionPermission = "crud_order"
	RUOutletAction  ActionPermission = "ru_outlet" */
	OrderAction ActionPermission = "order"

	// PRODUCT MANAGEMENT ACTIONS (PRODUCT CATEGORY, PRODUCT SUB CATEGORY, PRODUCT, STOCKS, RECIPE, INGREDIENT)
	CreateProductAction ActionPermission = "create_product"
	ReadProductAction   ActionPermission = "read_product"
	UpdateProductAction ActionPermission = "update_product"
	DeleteProductAction ActionPermission = "delete_product"

	CreateIngredientAction ActionPermission = "create_ingredient"
	ReadIngredientAction   ActionPermission = "read_ingredient"
	UpdateIngredientAction ActionPermission = "update_ingredient"
	DeleteIngredientAction ActionPermission = "delete_ingredient"

	ReadStockAction   ActionPermission = "read_stock"
	UpdateStockAction ActionPermission = "update_stock"

	// MARKETING MANAGEMENT ACTIONS (MEMBERSHIP, POINTS RULES, REDEMPTION RULES, VOUCHER)
	CreateMembershipAction ActionPermission = "create_membership"
	ReadMembershipAction   ActionPermission = "read_membership"
	UpdateMembershipAction ActionPermission = "update_membership"
	DeleteMembershipAction ActionPermission = "delete_membership"

	CreateVoucherAction ActionPermission = "create_voucher"
	ReadVoucherAction   ActionPermission = "read_voucher"
	UpdateVoucherAction ActionPermission = "update_voucher"
	DeleteVoucherAction ActionPermission = "delete_voucher"

	CreateDiscountAction ActionPermission = "create_discount"
	ReadDiscountAction   ActionPermission = "read_discount"
	UpdateDiscountAction ActionPermission = "update_discount"
	DeleteDiscountAction ActionPermission = "delete_discount"

	CreatePointsRulesAction ActionPermission = "create_points_rules"
	ReadPointsRulesAction   ActionPermission = "read_points_rules"
	UpdatePointsRulesAction ActionPermission = "update_points_rules"
	DeletePointsRulesAction ActionPermission = "delete_points_rules"

	CreateMissionAction ActionPermission = "create_mission"
	ReadMissionAction   ActionPermission = "read_mission"
	UpdateMissionAction ActionPermission = "update_mission"
	DeleteMissionAction ActionPermission = "delete_mission"

	// FINANCE MANAGEMENT ACTIONS (FINANCE)
	ReadFinanceOverviewAction     ActionPermission = "read_finance_overview"
	ReadFinanceFullReportAction   ActionPermission = "read_finance_full_report"
	ReadFinancePayoutReportAction ActionPermission = "read_finance_payout_report"
	ReadFinanceTransactionAction  ActionPermission = "read_finance_transaction"

	CRURedemptionRulesAction ActionPermission = "cru_redemption_rules"
	DRedemptionRulesAction   ActionPermission = "d_redemption_rules"

	// COMMUNICATION MANAGEMENT ACTIONS
	ManageOnboardingAction   ActionPermission = "manage_onboarding"
	ManageAnnouncementAction ActionPermission = "manage_announcement"
	ManageOrderMethodAction  ActionPermission = "manage_order_method"
	ManageFeedbackAction     ActionPermission = "manage_feedback"

	// SYSTEM MANAGEMENT ACTIONS (any actions related to system)
	CreateAppVersionAction ActionPermission = "create_app_version"
	ReadAppVersionAction   ActionPermission = "read_app_version"
	UpdateAppVersionAction ActionPermission = "update_app_version"
	DeleteAppVersionAction ActionPermission = "delete_app_version"

	// CAMPAIGN MANAGEMENT ACTIONS
	ManageCampaignAction ActionPermission = "manage_campaign"

	// POS MANAGEMENT ACTIONS (All actions which need RBAC within POS action)
	OutletOpenAction              ActionPermission = "outlet_open_action"         // outlet opening
	OutletCloseAction             ActionPermission = "outlet_close_action"        // outlet closing
	OutletToggleOnlineOrderAction ActionPermission = "outlet_toggle_online_order" // toggle online order
	CancelOrderAction             ActionPermission = "cancel_order_action"        // cancel order
	RefundOrderAction             ActionPermission = "refund_order_action"        // refund order
	CreateExpenseAction           ActionPermission = "create_expense_action"      // create expenses
	ReadExpenseAction             ActionPermission = "read_expense_action"        // read expenses
	ModifyExpenseAction           ActionPermission = "modify_expense_action"      // modify expenses
	POSInventoryAction            ActionPermission = "pos_inventory_action"       // pos inventory include all stock reports
	BindPosDeviceAction           ActionPermission = "pos_bind_device_action"     // bind pos device to outlet for cloud ECR
	POSSettingsAction             ActionPermission = "pos_settings_action"        // pos settings include all settings
)

type ModulePermission string

const (
	UserManagementModule              ModulePermission = "User Management"
	BusinessManagementModule          ModulePermission = "Business Management"
	OutletManagementModule            ModulePermission = "Outlet Management"
	ProductManagementModule           ModulePermission = "Product Management"
	MembershipLoyaltyManagementModule ModulePermission = "Membership Loyalty Management"
	FinanceManagementModule           ModulePermission = "Finance Management"
	CommunicationManagementModule     ModulePermission = "Communication Management"
	SystemManagementModule            ModulePermission = "System Management"
	CampaignManagementModule          ModulePermission = "Campaign Management"
	PosManagementModule               ModulePermission = "POS Management"
)

type SubModulePermission string

const (
	UserAdministrationSubModule        SubModulePermission = "User Administration"
	RoleManagementSubModule            SubModulePermission = "Role Management"
	AccessControlSubModule             SubModulePermission = "Access Control"
	BusinessCreationSubModule          SubModulePermission = "Business Creation"
	BusinessManagementSubModule        SubModulePermission = "Business Management"
	OutletCreationSubModule            SubModulePermission = "Outlet Creation"
	OutletManagementSubModule          SubModulePermission = "Outlet Management"
	OrderManagementSubModule           SubModulePermission = "Order Management"
	ProductManagementSubModule         SubModulePermission = "Product Management"
	MembershipManagementSubModule      SubModulePermission = "Membership Management"
	PointsRulesManagementSubModule     SubModulePermission = "Points Rules Management"
	RedemptionRulesManagementSubModule SubModulePermission = "Redemption Rules Management"
	VoucherManagementSubModule         SubModulePermission = "Voucher Management"
	DiscountManagementSubModule        SubModulePermission = "Discount Management"
	MissionManagementSubModule         SubModulePermission = "Mission Management"
	IngredientManagementSubModule      SubModulePermission = "Ingredient Management"
	StockManagementSubModule           SubModulePermission = "Stock Management"
	FinanceManagementSubModule         SubModulePermission = "Overview"
	FinanceFullReportSubModule         SubModulePermission = "Full Report"
	FinancePayoutReportSubModule       SubModulePermission = "Payout Report"
	FinanceTransactionSubModule        SubModulePermission = "Transaction"
	OnboardingSubModule                SubModulePermission = "Onboarding"
	AnnouncementSubModule              SubModulePermission = "Announcement"
	OrderMethodSubModule               SubModulePermission = "Order Method"
	FeedbackSubModule                  SubModulePermission = "Feedback"
	AppManagementSubModule             SubModulePermission = "App Management"
	CampaignManagementSubModule        SubModulePermission = "Campaign Notification Management"
	POSOutletManagementSubModule       SubModulePermission = "POS Outlet Management"
)

type PermissionPresetFormat struct {
	ID              uuid.UUID
	GroupRoleID     uuid.UUID
	Type            string
	Name            ActionPermission
	Description     string
	Module          ModulePermission
	SubModule       SubModulePermission
	ModuleOrder     int
	SubModuleOrder  int
	PermissionOrder int
	Enabled         bool
}

var ModuleOrders = map[ModulePermission]int{
	UserManagementModule:              1,
	BusinessManagementModule:          2,
	OutletManagementModule:            3,
	ProductManagementModule:           4,
	MembershipLoyaltyManagementModule: 5,
	FinanceManagementModule:           6,
	CommunicationManagementModule:     7,
	SystemManagementModule:            8,
	CampaignManagementModule:          9,
	PosManagementModule:               10,
}

var SubModuleOrders = map[SubModulePermission]int{
	UserAdministrationSubModule:    1,
	RoleManagementSubModule:        2,
	AccessControlSubModule:         3,
	BusinessCreationSubModule:      1,
	BusinessManagementSubModule:    2,
	OutletCreationSubModule:        1,
	OutletManagementSubModule:      2,
	OrderManagementSubModule:       3,
	ProductManagementSubModule:     1,
	IngredientManagementSubModule:  2,
	StockManagementSubModule:       3,
	MembershipManagementSubModule:  1,
	VoucherManagementSubModule:     2,
	DiscountManagementSubModule:    3,
	PointsRulesManagementSubModule: 4,
	MissionManagementSubModule:     5,
	FinanceManagementSubModule:     1,
	FinanceFullReportSubModule:     2,
	FinancePayoutReportSubModule:   3,
	FinanceTransactionSubModule:    4,
	OnboardingSubModule:            1,
	AnnouncementSubModule:          2,
	OrderMethodSubModule:           3,
	FeedbackSubModule:              4,
	AppManagementSubModule:         1,
	CampaignManagementSubModule:    1,
	POSOutletManagementSubModule:   1,
}

var PermissionPresetsArray = []PermissionPresetFormat{
	{
		Name:            CreateUserAction,
		Description:     "Create user accounts",
		Module:          UserManagementModule,
		SubModule:       UserAdministrationSubModule,
		ModuleOrder:     ModuleOrders[UserManagementModule],
		SubModuleOrder:  SubModuleOrders[UserAdministrationSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ReadUserAction,
		Description:     "Read user accounts",
		Module:          UserManagementModule,
		SubModule:       UserAdministrationSubModule,
		ModuleOrder:     ModuleOrders[UserManagementModule],
		SubModuleOrder:  SubModuleOrders[UserAdministrationSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            UpdateUserAction,
		Description:     "Update user accounts",
		Module:          UserManagementModule,
		SubModule:       UserAdministrationSubModule,
		ModuleOrder:     ModuleOrders[UserManagementModule],
		SubModuleOrder:  SubModuleOrders[UserAdministrationSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            DeleteUserAction,
		Description:     "Delete user accounts",
		Module:          UserManagementModule,
		SubModule:       UserAdministrationSubModule,
		ModuleOrder:     ModuleOrders[UserManagementModule],
		SubModuleOrder:  SubModuleOrders[UserAdministrationSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            ResetPasswordAction,
		Description:     "Enable password reset and recovery options",
		Module:          UserManagementModule,
		SubModule:       UserAdministrationSubModule,
		ModuleOrder:     ModuleOrders[UserManagementModule],
		SubModuleOrder:  SubModuleOrders[UserAdministrationSubModule],
		PermissionOrder: 5,
		Enabled:         false,
	},
	{
		Name:            CreateRoleAction,
		Description:     "Create role",
		Module:          UserManagementModule,
		SubModule:       RoleManagementSubModule,
		ModuleOrder:     ModuleOrders[UserManagementModule],
		SubModuleOrder:  SubModuleOrders[RoleManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ReadRoleAction,
		Description:     "Read role",
		Module:          UserManagementModule,
		SubModule:       RoleManagementSubModule,
		ModuleOrder:     ModuleOrders[UserManagementModule],
		SubModuleOrder:  SubModuleOrders[RoleManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            UpdateRoleAction,
		Description:     "Update role",
		Module:          UserManagementModule,
		SubModule:       RoleManagementSubModule,
		ModuleOrder:     ModuleOrders[UserManagementModule],
		SubModuleOrder:  SubModuleOrders[RoleManagementSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            DeleteRoleAction,
		Description:     "Delete role",
		Module:          UserManagementModule,
		SubModule:       RoleManagementSubModule,
		ModuleOrder:     ModuleOrders[UserManagementModule],
		SubModuleOrder:  SubModuleOrders[RoleManagementSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            UserPermissionAction,
		Description:     "User Permissions",
		Module:          UserManagementModule,
		SubModule:       AccessControlSubModule,
		ModuleOrder:     ModuleOrders[UserManagementModule],
		SubModuleOrder:  SubModuleOrders[AccessControlSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ActivityLogAction,
		Description:     "View activity logging",
		Module:          UserManagementModule,
		SubModule:       AccessControlSubModule,
		ModuleOrder:     ModuleOrders[UserManagementModule],
		SubModuleOrder:  SubModuleOrders[AccessControlSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            CreateBusinessAction,
		Description:     "Create business account",
		Module:          BusinessManagementModule,
		SubModule:       BusinessCreationSubModule,
		ModuleOrder:     ModuleOrders[BusinessManagementModule],
		SubModuleOrder:  SubModuleOrders[BusinessCreationSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            DeleteBusinessAction,
		Description:     "Delete business account",
		Module:          BusinessManagementModule,
		SubModule:       BusinessCreationSubModule,
		ModuleOrder:     ModuleOrders[BusinessManagementModule],
		SubModuleOrder:  SubModuleOrders[BusinessCreationSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            ReadBusinessAction,
		Description:     "Read business account",
		Module:          BusinessManagementModule,
		SubModule:       BusinessManagementSubModule,
		ModuleOrder:     ModuleOrders[BusinessManagementModule],
		SubModuleOrder:  SubModuleOrders[BusinessManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            UpdateBusinessAction,
		Description:     "Update business account",
		Module:          BusinessManagementModule,
		SubModule:       BusinessManagementSubModule,
		ModuleOrder:     ModuleOrders[BusinessManagementModule],
		SubModuleOrder:  SubModuleOrders[BusinessManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            CreateOutletAction,
		Description:     "Create outlet",
		Module:          OutletManagementModule,
		SubModule:       OutletCreationSubModule,
		ModuleOrder:     ModuleOrders[OutletManagementModule],
		SubModuleOrder:  SubModuleOrders[OutletCreationSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            DeleteOutletAction,
		Description:     "Delete outlet",
		Module:          OutletManagementModule,
		SubModule:       OutletCreationSubModule,
		ModuleOrder:     ModuleOrders[OutletManagementModule],
		SubModuleOrder:  SubModuleOrders[OutletCreationSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            ReadOutletAction,
		Description:     "Read outlet",
		Module:          OutletManagementModule,
		SubModule:       OutletManagementSubModule,
		ModuleOrder:     ModuleOrders[OutletManagementModule],
		SubModuleOrder:  SubModuleOrders[OutletManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            UpdateOutletAction,
		Description:     "Update outlet",
		Module:          OutletManagementModule,
		SubModule:       OutletManagementSubModule,
		ModuleOrder:     ModuleOrders[OutletManagementModule],
		SubModuleOrder:  SubModuleOrders[OutletManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            OrderAction,
		Description:     "Order Action",
		Module:          OutletManagementModule,
		SubModule:       OrderManagementSubModule,
		ModuleOrder:     ModuleOrders[OutletManagementModule],
		SubModuleOrder:  SubModuleOrders[OrderManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            CreateProductAction,
		Description:     "Create products and modifiers",
		Module:          ProductManagementModule,
		SubModule:       ProductManagementSubModule,
		ModuleOrder:     ModuleOrders[ProductManagementModule],
		SubModuleOrder:  SubModuleOrders[ProductManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ReadProductAction,
		Description:     "Read products and modifiers",
		Module:          ProductManagementModule,
		SubModule:       ProductManagementSubModule,
		ModuleOrder:     ModuleOrders[ProductManagementModule],
		SubModuleOrder:  SubModuleOrders[ProductManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            UpdateProductAction,
		Description:     "Update products and modifiers",
		Module:          ProductManagementModule,
		SubModule:       ProductManagementSubModule,
		ModuleOrder:     ModuleOrders[ProductManagementModule],
		SubModuleOrder:  SubModuleOrders[ProductManagementSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            DeleteProductAction,
		Description:     "Delete products and modifiers",
		Module:          ProductManagementModule,
		SubModule:       ProductManagementSubModule,
		ModuleOrder:     ModuleOrders[ProductManagementModule],
		SubModuleOrder:  SubModuleOrders[ProductManagementSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            CreateIngredientAction,
		Description:     "Create ingredients and recipes",
		Module:          ProductManagementModule,
		SubModule:       IngredientManagementSubModule,
		ModuleOrder:     ModuleOrders[ProductManagementModule],
		SubModuleOrder:  SubModuleOrders[IngredientManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ReadIngredientAction,
		Description:     "Read ingredients and recipes",
		Module:          ProductManagementModule,
		SubModule:       IngredientManagementSubModule,
		ModuleOrder:     ModuleOrders[ProductManagementModule],
		SubModuleOrder:  SubModuleOrders[IngredientManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            UpdateIngredientAction,
		Description:     "Update ingredients and recipes",
		Module:          ProductManagementModule,
		SubModule:       IngredientManagementSubModule,
		ModuleOrder:     ModuleOrders[ProductManagementModule],
		SubModuleOrder:  SubModuleOrders[IngredientManagementSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            DeleteIngredientAction,
		Description:     "Delete ingredients and recipes",
		Module:          ProductManagementModule,
		SubModule:       IngredientManagementSubModule,
		ModuleOrder:     ModuleOrders[ProductManagementModule],
		SubModuleOrder:  SubModuleOrders[IngredientManagementSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            ReadStockAction,
		Description:     "Read stocks",
		Module:          ProductManagementModule,
		SubModule:       StockManagementSubModule,
		ModuleOrder:     ModuleOrders[ProductManagementModule],
		SubModuleOrder:  SubModuleOrders[StockManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            UpdateStockAction,
		Description:     "Update stocks",
		Module:          ProductManagementModule,
		SubModule:       StockManagementSubModule,
		ModuleOrder:     ModuleOrders[ProductManagementModule],
		SubModuleOrder:  SubModuleOrders[StockManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            CreateMembershipAction,
		Description:     "Create membership",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       MembershipManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[MembershipManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ReadMembershipAction,
		Description:     "Read membership",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       MembershipManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[MembershipManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            UpdateMembershipAction,
		Description:     "Update membership",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       MembershipManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[MembershipManagementSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            DeleteMembershipAction,
		Description:     "Delete membership",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       MembershipManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[MembershipManagementSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            CreateVoucherAction,
		Description:     "Create vouchers",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       VoucherManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[VoucherManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ReadVoucherAction,
		Description:     "Read vouchers",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       VoucherManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[VoucherManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            UpdateVoucherAction,
		Description:     "Update vouchers",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       VoucherManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[VoucherManagementSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            DeleteVoucherAction,
		Description:     "Delete vouchers",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       VoucherManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[VoucherManagementSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            CreateDiscountAction,
		Description:     "Create discount",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       DiscountManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[DiscountManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ReadDiscountAction,
		Description:     "Read discount",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       DiscountManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[DiscountManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            UpdateDiscountAction,
		Description:     "Update discount",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       DiscountManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[DiscountManagementSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            DeleteDiscountAction,
		Description:     "Delete discount",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       DiscountManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[DiscountManagementSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            CreatePointsRulesAction,
		Description:     "Create points rules",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       PointsRulesManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[PointsRulesManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ReadPointsRulesAction,
		Description:     "Read points rules",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       PointsRulesManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[PointsRulesManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            UpdatePointsRulesAction,
		Description:     "Update points rules",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       PointsRulesManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[PointsRulesManagementSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            DeletePointsRulesAction,
		Description:     "Delete points rules",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       PointsRulesManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[PointsRulesManagementSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            CreateMissionAction,
		Description:     "Create mission",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       MissionManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[MissionManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ReadMissionAction,
		Description:     "Read mission",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       MissionManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[MissionManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            UpdateMissionAction,
		Description:     "Update mission",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       MissionManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[MissionManagementSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            DeleteMissionAction,
		Description:     "Delete mission",
		Module:          MembershipLoyaltyManagementModule,
		SubModule:       MissionManagementSubModule,
		ModuleOrder:     ModuleOrders[MembershipLoyaltyManagementModule],
		SubModuleOrder:  SubModuleOrders[MissionManagementSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            ReadFinanceOverviewAction,
		Description:     "Read finance overview",
		Module:          FinanceManagementModule,
		SubModule:       FinanceManagementSubModule,
		ModuleOrder:     ModuleOrders[FinanceManagementModule],
		SubModuleOrder:  SubModuleOrders[FinanceManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ReadFinanceFullReportAction,
		Description:     "Read finance full report",
		Module:          FinanceManagementModule,
		SubModule:       FinanceFullReportSubModule,
		ModuleOrder:     ModuleOrders[FinanceManagementModule],
		SubModuleOrder:  SubModuleOrders[FinanceFullReportSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            ReadFinancePayoutReportAction,
		Description:     "Read finance payout report",
		Module:          FinanceManagementModule,
		SubModule:       FinancePayoutReportSubModule,
		ModuleOrder:     ModuleOrders[FinanceManagementModule],
		SubModuleOrder:  SubModuleOrders[FinancePayoutReportSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            ReadFinanceTransactionAction,
		Description:     "Read finance transaction",
		Module:          FinanceManagementModule,
		SubModule:       FinanceTransactionSubModule,
		ModuleOrder:     ModuleOrders[FinanceManagementModule],
		SubModuleOrder:  SubModuleOrders[FinanceTransactionSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            ManageOnboardingAction,
		Description:     "Manage onboarding",
		Module:          CommunicationManagementModule,
		SubModule:       OnboardingSubModule,
		ModuleOrder:     ModuleOrders[CommunicationManagementModule],
		SubModuleOrder:  SubModuleOrders[OnboardingSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ManageAnnouncementAction,
		Description:     "Manage announcement",
		Module:          CommunicationManagementModule,
		SubModule:       AnnouncementSubModule,
		ModuleOrder:     ModuleOrders[CommunicationManagementModule],
		SubModuleOrder:  SubModuleOrders[AnnouncementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            ManageOrderMethodAction,
		Description:     "Manage order method",
		Module:          CommunicationManagementModule,
		SubModule:       OrderMethodSubModule,
		ModuleOrder:     ModuleOrders[CommunicationManagementModule],
		SubModuleOrder:  SubModuleOrders[OrderMethodSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            ManageFeedbackAction,
		Description:     "Manage feedback",
		Module:          CommunicationManagementModule,
		SubModule:       FeedbackSubModule,
		ModuleOrder:     ModuleOrders[CommunicationManagementModule],
		SubModuleOrder:  SubModuleOrders[FeedbackSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            CreateAppVersionAction,
		Description:     "Create app version",
		Module:          SystemManagementModule,
		SubModule:       AppManagementSubModule,
		ModuleOrder:     ModuleOrders[SystemManagementModule],
		SubModuleOrder:  SubModuleOrders[AppManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            ReadAppVersionAction,
		Description:     "Read app version",
		Module:          SystemManagementModule,
		SubModule:       AppManagementSubModule,
		ModuleOrder:     ModuleOrders[SystemManagementModule],
		SubModuleOrder:  SubModuleOrders[AppManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            UpdateAppVersionAction,
		Description:     "Update app version",
		Module:          SystemManagementModule,
		SubModule:       AppManagementSubModule,
		ModuleOrder:     ModuleOrders[SystemManagementModule],
		SubModuleOrder:  SubModuleOrders[AppManagementSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            DeleteAppVersionAction,
		Description:     "Delete app version",
		Module:          SystemManagementModule,
		SubModule:       AppManagementSubModule,
		ModuleOrder:     ModuleOrders[SystemManagementModule],
		SubModuleOrder:  SubModuleOrders[AppManagementSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            ManageCampaignAction,
		Description:     "Manage campaign notification",
		Module:          CampaignManagementModule,
		SubModule:       CampaignManagementSubModule,
		ModuleOrder:     ModuleOrders[CampaignManagementModule],
		SubModuleOrder:  SubModuleOrders[CampaignManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            OutletOpenAction,
		Description:     "Open POS outlet",
		Module:          PosManagementModule,
		SubModule:       POSOutletManagementSubModule,
		ModuleOrder:     ModuleOrders[PosManagementModule],
		SubModuleOrder:  SubModuleOrders[POSOutletManagementSubModule],
		PermissionOrder: 1,
		Enabled:         false,
	},
	{
		Name:            OutletCloseAction,
		Description:     "Close POS outlet",
		Module:          PosManagementModule,
		SubModule:       POSOutletManagementSubModule,
		ModuleOrder:     ModuleOrders[PosManagementModule],
		SubModuleOrder:  SubModuleOrders[POSOutletManagementSubModule],
		PermissionOrder: 2,
		Enabled:         false,
	},
	{
		Name:            OutletToggleOnlineOrderAction,
		Description:     "Toggle online order for POS outlet",
		Module:          PosManagementModule,
		SubModule:       POSOutletManagementSubModule,
		ModuleOrder:     ModuleOrders[PosManagementModule],
		SubModuleOrder:  SubModuleOrders[POSOutletManagementSubModule],
		PermissionOrder: 3,
		Enabled:         false,
	},
	{
		Name:            CancelOrderAction,
		Description:     "Cancel order for POS outlet",
		Module:          PosManagementModule,
		SubModule:       POSOutletManagementSubModule,
		ModuleOrder:     ModuleOrders[PosManagementModule],
		SubModuleOrder:  SubModuleOrders[POSOutletManagementSubModule],
		PermissionOrder: 4,
		Enabled:         false,
	},
	{
		Name:            RefundOrderAction,
		Description:     "Refund order for POS outlet",
		Module:          PosManagementModule,
		SubModule:       POSOutletManagementSubModule,
		ModuleOrder:     ModuleOrders[PosManagementModule],
		SubModuleOrder:  SubModuleOrders[POSOutletManagementSubModule],
		PermissionOrder: 5,
		Enabled:         false,
	},
	{
		Name:            CreateExpenseAction,
		Description:     "Create expense for POS outlet",
		Module:          PosManagementModule,
		SubModule:       POSOutletManagementSubModule,
		ModuleOrder:     ModuleOrders[PosManagementModule],
		SubModuleOrder:  SubModuleOrders[POSOutletManagementSubModule],
		PermissionOrder: 6,
		Enabled:         false,
	},
	{
		Name:            ReadExpenseAction,
		Description:     "Read expense for POS outlet",
		Module:          PosManagementModule,
		SubModule:       POSOutletManagementSubModule,
		ModuleOrder:     ModuleOrders[PosManagementModule],
		SubModuleOrder:  SubModuleOrders[POSOutletManagementSubModule],
		PermissionOrder: 7,
		Enabled:         false,
	},
	{
		Name:            ModifyExpenseAction,
		Description:     "Modify expense for POS outlet",
		Module:          PosManagementModule,
		SubModule:       POSOutletManagementSubModule,
		ModuleOrder:     ModuleOrders[PosManagementModule],
		SubModuleOrder:  SubModuleOrders[POSOutletManagementSubModule],
		PermissionOrder: 8,
		Enabled:         false,
	},
	{
		Name:            POSInventoryAction,
		Description:     "POS inventory (stock reports) for outlet",
		Module:          PosManagementModule,
		SubModule:       POSOutletManagementSubModule,
		ModuleOrder:     ModuleOrders[PosManagementModule],
		SubModuleOrder:  SubModuleOrders[POSOutletManagementSubModule],
		PermissionOrder: 9,
		Enabled:         false,
	},
	{
		Name:            BindPosDeviceAction,
		Description:     "Bind POS device to outlet (cloud ECR)",
		Module:          PosManagementModule,
		SubModule:       POSOutletManagementSubModule,
		ModuleOrder:     ModuleOrders[PosManagementModule],
		SubModuleOrder:  SubModuleOrders[POSOutletManagementSubModule],
		PermissionOrder: 10,
		Enabled:         false,
	},
	{
		Name:            POSSettingsAction,
		Description:     "POS settings for outlet",
		Module:          PosManagementModule,
		SubModule:       POSOutletManagementSubModule,
		ModuleOrder:     ModuleOrders[PosManagementModule],
		SubModuleOrder:  SubModuleOrders[POSOutletManagementSubModule],
		PermissionOrder: 11,
		Enabled:         false,
	},
}
