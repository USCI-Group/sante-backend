package database

import (
	"errors"

	"encore.app/common/constants"
	"encore.app/database/models"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

var santeDB = sqldb.NewDatabase("sante", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// initService initializes the site service.
// It is automatically called by Encore on service startup.
func initService() (*Service, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: santeDB.Stdlib(),
	}))
	if err != nil {
		return nil, err
	}

	if err := seedDatabase(db); err != nil {
		return nil, err
	}

	return &Service{db: db}, nil
}

func GetSanteDB() *sqldb.Database {
	return santeDB
}

// Seeders

func seedRoles(db *gorm.DB) error {
	roles := constants.UserRoles

	for _, roleName := range roles {
		var existingRole models.Role
		if err := db.Where("name = ?", roleName).First(&existingRole).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newRole := &models.Role{
					Name:     string(roleName),
					RoleType: constants.RoleTypeGeneral,
				}
				if roleName == constants.SanteSuperAdminRole {
					newRole.RoleType = constants.RoleTypeAdmin
				}
				if err := db.Create(newRole).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}

func seedDatabase(db *gorm.DB) error {
	if err := seedRoles(db); err != nil {
		return err
	}

	if err := seedSanteAdminGroupRoles(db); err != nil {
		return err
	}

	if err := seedSuperSanteAdminPermissions(db); err != nil {
		return err
	}

	if err := seedOfficialMenu(db); err != nil {
		return err
	}

	return nil
}

func seedOfficialMenu(db *gorm.DB) error {
	bizID, _ := uuid.FromString("7d42f515-3517-4e76-be13-30880443546f")
	
	categories := []models.ProductCategory{
		{Name: "Burgers", Description: "Gourmet healthy burgers"},
		{Name: "Wraps", Description: "Fresh wholemeal wraps"},
		{Name: "Rice Bowls", Description: "Healthy grain bowls"},
		{Name: "Sandwiches", Description: "Wholemeal artisan sandwiches"},
		{Name: "Beverages", Description: "Healthy refreshments"},
		{Name: "Bakery & Desserts", Description: "Freshly baked goods"},
	}

	for i, cat := range categories {
		var existing models.ProductCategory
		if err := db.Where("name = ? AND business_id = ?", cat.Name, bizID).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				cat.BusinessID = &bizID
				cat.SortOrder = i
				if err := db.Create(&cat).Error; err != nil {
					return err
				}
				existing = cat
			}
		}

		// Seed some products for each category
		if cat.Name == "Burgers" {
			products := []models.Product{
				{Name: "Beefy Mushroom Melt", Price: 27.90, Description: "100% Aussie beef patty with melted cheddar, lettuce and pickles, served with creamy mushroom sauce on the side."},
				{Name: "The Santé Clucker", Price: 13.90, Description: "Grilled chicken thigh topped with crisp lettuce, tomato, and pickles, served with a side of our signature Santé sauce."},
				{Name: "Thai-namite Stack", Price: 26.90, Description: "100% Aussie beef patty layered with fresh lettuce, tomato, spicy jalapeño and caramelised onions, with a side of spicy green mayo."},
			}
			for _, p := range products {
				_seedProduct(db, p, existing.ID, bizID)
			}
		} else if cat.Name == "Rice Bowls" {
			products := []models.Product{
				{Name: "Thai Spice Shrimp Bowl", Price: 20.90, Description: "Succulent shrimp served on brown rice with crisp lettuce and pickles, topped with Japanese furikake seasoning and spicy green Thai sauce."},
				{Name: "Tuna Topper Rice Bowl", Price: 22.90, Description: "Tender tuna chunks over brown rice mixed with fresh lettuce, garnished with green onions and sesame seeds drizzled with roasted sesame sauce."},
			}
			for _, p := range products {
				_seedProduct(db, p, existing.ID, bizID)
			}
		} else if cat.Name == "Beverages" {
			products := []models.Product{
				{Name: "Roselle Spritz", Price: 4.90, Description: "Refreshing roselle flavored sparkling water."},
				{Name: "Americano (M)", Price: 5.00, Description: "Rich classic black coffee (Medium)."},
				{Name: "Orange Sunburst", Price: 13.50, Description: "Fresh orange citrus refreshment."},
			}
			for _, p := range products {
				_seedProduct(db, p, existing.ID, bizID)
			}
		}
	}
	return nil
}

func _seedProduct(db *gorm.DB, p models.Product, catID uuid.UUID, bizID uuid.UUID) {
	var existing models.Product
	if err := db.Where("name = ? AND business_id = ?", p.Name, bizID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			p.BusinessID = &bizID
			p.IsActive = true
			p.IsStoreOutlet = true
			if err := db.Create(&p).Error; err != nil {
				return
			}
			// Map to category
			mapping := models.ProductCategoryMapping{
				ProductID:         p.ID,
				ProductCategoryID: catID,
			}
			db.Create(&mapping)
		}
	}
}

func seedSanteAdminGroupRoles(db *gorm.DB) error {
	// Create group role for sante super admin
	var superAdminRole models.Role
	if err := db.Where("name = ?", constants.SanteSuperAdminRole).First(&superAdminRole).Error; err != nil {
		return err
	}

	var existingSuperAdminGroupRole models.GroupRole
	if err := db.Where("business_id IS NULL AND role_id = ?", superAdminRole.ID).First(&existingSuperAdminGroupRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newGroupRole := &models.GroupRole{
				RoleID: superAdminRole.ID,
			}
			if err := db.Create(newGroupRole).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Create super admin user if not exists
	var existingSuperAdmin models.User
	if err := db.Where("email = ?", "super.admin@sante.com").First(&existingSuperAdmin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var superAdminGroupRole models.GroupRole
			if err := db.Joins("JOIN roles ON roles.id = group_roles.role_id").
				Where("roles.name = ? AND group_roles.business_id IS NULL", "sante_super_admin").
				First(&superAdminGroupRole).Error; err != nil {
				return err
			}

			superAdmin := &models.User{
				Email:       "super.admin@sante.com",
				Pwd:         "$2a$10$fIISvPdUGPafU5KPlYok1.mYWjaqk2Lw18HnEFZg2.NpwXrhVgbnS",
				FirstName:   "Super Admin",
				Surname:     "Sante",
				GroupRoleID: &superAdminGroupRole.ID,
			}
			if err := db.Create(superAdmin).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func seedSuperSanteAdminPermissions(db *gorm.DB) error {
	// seed permissions for super sante admin
	var sante_super_admin models.Role
	if err := db.Where("name = ?", constants.SanteSuperAdminRole).First(&sante_super_admin).Error; err != nil {
		return err
	}

	var sante_super_admin_group_role models.GroupRole
	if err := db.Where("role_id = ? AND business_id IS NULL", sante_super_admin.ID).First(&sante_super_admin_group_role).Error; err != nil {
		return err
	}

	// Add all permissions to sante super admin group role if they don't exist
	for _, p := range constants.PermissionPresetsArray {
		// Add Permission Presets
		var existingPermissionPreset models.PermissionPreset
		if err := db.Where("name = ? AND role_id = ?", p.Name, sante_super_admin.ID).First(&existingPermissionPreset).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newPermissionPreset := &models.PermissionPreset{
					Name:      p.Name,
					Module:    p.Module,
					SubModule: p.SubModule,
					RoleID:    sante_super_admin.ID,
					Enabled:   true,
				}
				if err := db.Create(newPermissionPreset).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// Add Permissions
		var existingPermission models.Permission
		if err := db.Where("name = ? AND group_role_id = ?", p.Name, sante_super_admin_group_role.ID).First(&existingPermission).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newPermission := &models.Permission{
					Name:        p.Name,
					Module:      p.Module,
					SubModule:   p.SubModule,
					GroupRoleID: sante_super_admin_group_role.ID,
					Enabled:     true,
				}
				if err := db.Create(newPermission).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}
