package constants

type UserRole string

const (
	SanteSuperAdminRole UserRole = "sante_super_admin"
	// BusinessOwnerRole    UserRole = "business_owner"
	// BusinessAdminRole    UserRole = "business_admin"
	// OutletAdminRole      UserRole = "outlet_admin"
	// OutletManagerRole    UserRole = "outlet_manager"
	// OutletStaffRole      UserRole = "staff"
	// OutletCashierRole    UserRole = "cashier"
	// OutletWaiterRole     UserRole = "waiter"
)

type RoleType string

const (
	RoleTypeAdmin    RoleType = "admin"
	RoleTypeGeneral  RoleType = "general"
	RoleTypeBusiness RoleType = "business"
)

var SanteAdminRoles = []UserRole{
	SanteSuperAdminRole,
}
var BusinessPermissionRoles = []UserRole{
	SanteSuperAdminRole,
	// BusinessOwnerRole,
	// BusinessAdminRole,
}

var UserRoles = []UserRole{
	SanteSuperAdminRole,
	// BusinessOwnerRole,
	// BusinessAdminRole,
	// OutletAdminRole,
	// OutletManagerRole,
	// OutletStaffRole,
	// OutletCashierRole,
	// OutletWaiterRole,
}
