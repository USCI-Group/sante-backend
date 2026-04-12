package businesses

import (
	"context"
	"errors"
	"log"

	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/database/models"
	"encore.app/middleware"
	"encore.dev/types/uuid"
	"github.com/asaskevich/govalidator"
	"gorm.io/gorm"
)

type SaveBusinessConfigurationParams struct {
	BusinessID              uuid.UUID `json:"business_id"`
	GrabClientID            string    `json:"grab_client_id"`
	GrabClientSecret        string    `json:"grab_client_secret"`
	GrabExpressClientID     string    `json:"grab_express_client_id"`
	GrabExpressClientSecret string    `json:"grab_express_client_secret"`
	ShopeeClientID          string    `json:"shopee_client_id"`
	ShopeeClientSecret      string    `json:"shopee_client_secret"`
	ServiceCharge           *float32  `json:"service_charge"`
	ServiceTax              *float32  `json:"service_tax"`
}

type GetLogoutButtonVisibilityResponse struct {
	Message string `json:"message"`
	Data    bool   `json:"data"`
}

type TermsOfServiceParams struct {
	BusinessID     uuid.UUID `json:"business_id"`
	TermsOfService string    `json:"terms_of_service"`
}

type PrivacyPolicyParams struct {
	BusinessID    uuid.UUID `json:"business_id"`
	PrivacyPolicy string    `json:"privacy_policy"`
}

type GetTnCAndPrivacyPolicyResponse struct {
	TermsOfService string `json:"terms_of_service"`
	PrivacyPolicy  string `json:"privacy_policy"`
}

// encore:api auth method=POST path=/api/admin/business/configuration/save
func (s *Service) SaveBusinessConfiguration(ctx context.Context, params *SaveBusinessConfigurationParams) (*models.BusinessConfiguration, error) {
	if err := middleware.CheckSanteAdmin(); err != nil {
		return nil, err
	}

	if _, err := govalidator.ValidateStruct(params); err != nil {
		return nil, err
	}

	var business models.Business
	if err := s.db.First(&business, "id = ?", params.BusinessID).Error; err != nil {
		return nil, err
	}

	var grabClientSecret string
	var grabExpressClientSecret string
	var shopeeClientSecret string

	var err error
	if params.GrabClientSecret != "" {
		grabClientSecret, err = common.EncryptText(params.GrabClientSecret)
		if err != nil {
			return nil, err
		}
	}

	if params.GrabExpressClientSecret != "" {
		grabExpressClientSecret, err = common.EncryptText(params.GrabExpressClientSecret)
		if err != nil {
			return nil, err
		}
	}

	if params.ShopeeClientSecret != "" {
		shopeeClientSecret, err = common.EncryptText(params.ShopeeClientSecret)
		if err != nil {
			return nil, err
		}
	}

	var businessConfiguration models.BusinessConfiguration
	if err := s.db.First(&businessConfiguration, "business_id = ?", params.BusinessID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			businessConfiguration = models.BusinessConfiguration{
				BusinessID:              params.BusinessID,
				GrabClientID:            &params.GrabClientID,
				GrabClientSecret:        &grabClientSecret,
				GrabExpressClientID:     &params.GrabExpressClientID,
				GrabExpressClientSecret: &grabExpressClientSecret,
				ShopeeClientID:          &params.ShopeeClientID,
				ShopeeClientSecret:      &shopeeClientSecret,
				ServiceChargePercentage: params.ServiceCharge,
				ServiceTaxPercentage:    params.ServiceTax,
			}

			if err := s.db.Create(&businessConfiguration).Error; err != nil {
				log.Printf("unable to create business configuration: %v", err)
				return nil, err
			}
		} else {
			return nil, err
		}

	} else {
		businessConfiguration.GrabClientID = &params.GrabClientID
		businessConfiguration.GrabExpressClientID = &params.GrabExpressClientID
		businessConfiguration.ShopeeClientID = &params.ShopeeClientID
		if grabClientSecret != "" {
			businessConfiguration.GrabClientSecret = &grabClientSecret
		}
		if grabExpressClientSecret != "" {
			businessConfiguration.GrabExpressClientSecret = &grabExpressClientSecret
		}
		if shopeeClientSecret != "" {
			businessConfiguration.ShopeeClientSecret = &shopeeClientSecret
		}
		if params.ServiceCharge != nil {
			businessConfiguration.ServiceChargePercentage = params.ServiceCharge
		}
		if params.ServiceTax != nil {
			businessConfiguration.ServiceTaxPercentage = params.ServiceTax
		}
		if err := s.db.Save(&businessConfiguration).Error; err != nil {
			return nil, err
		}
	}

	return &businessConfiguration, nil
}

//encore:api auth method=GET path=/api/admin/business/configuration/get/:id
func (s *Service) GetBusinessConfiguration(ctx context.Context, id uuid.UUID) (*models.BusinessConfiguration, error) {
	var businessConfiguration models.BusinessConfiguration
	if err := s.db.First(&businessConfiguration, "business_id = ?", id).Error; err != nil {
		return nil, err
	}
	return &businessConfiguration, nil
}

// API to get logout button visibility of a business
//
//encore:api auth method=GET path=/api/business/configuration/get/pos-logout-button/:business_id
func (s *Service) GetLogoutButtonVisibility(ctx context.Context, business_id uuid.UUID) (*GetLogoutButtonVisibilityResponse, error) {
	var businessConfig models.BusinessConfiguration
	result := s.db.Model(&models.BusinessConfiguration{}).Where("business_id = ?", business_id).First(&businessConfig)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetLogoutButtonVisibilityResponse{
		Message: "Logout button visibility retrieved successfully",
		Data:    businessConfig.IsLogoutButtonVisible,
	}, nil
}

// API to save terms of service of a business
//
//encore:api auth method=POST path=/api/admin/business/configuration/tnc/save
func (s *Service) SaveTermsOfService(ctx context.Context, params *TermsOfServiceParams) error {
	if err := middleware.CheckPermission(constants.UpdateBusinessAction, &params.BusinessID, nil); err != nil {
		return err
	}

	var businessConfig models.BusinessConfiguration
	result := s.db.Model(&models.BusinessConfiguration{}).Where("business_id = ?", params.BusinessID).First(&businessConfig)
	if result.Error != nil {
		return result.Error
	}

	businessConfig.TermsOfService = &params.TermsOfService
	if err := s.db.Save(&businessConfig).Error; err != nil {
		return err
	}

	return nil
}

// API to save privacy policy of a business
//
//encore:api auth method=POST path=/api/admin/business/configuration/privacy-policy/save
func (s *Service) SavePrivacyPolicy(ctx context.Context, params *PrivacyPolicyParams) error {
	if err := middleware.CheckPermission(constants.UpdateBusinessAction, &params.BusinessID, nil); err != nil {
		return err
	}

	var businessConfig models.BusinessConfiguration
	result := s.db.Model(&models.BusinessConfiguration{}).Where("business_id = ?", params.BusinessID).First(&businessConfig)
	if result.Error != nil {
		return result.Error
	}

	businessConfig.PrivacyPolicy = &params.PrivacyPolicy
	if err := s.db.Save(&businessConfig).Error; err != nil {
		return err
	}

	return nil
}

// API to get terms of service and privacy policy of a business
//
//encore:api public method=GET path=/api/business/configuration/tnc-and-privacy-policy/:business_id
func (s *Service) GetTnCAndPrivacyPolicy(ctx context.Context, business_id uuid.UUID) (*GetTnCAndPrivacyPolicyResponse, error) {
	var businessConfiguration models.BusinessConfiguration
	if err := s.db.First(&businessConfiguration, "business_id = ?", business_id).Error; err != nil {
		return nil, err
	}
	return &GetTnCAndPrivacyPolicyResponse{
		TermsOfService: *businessConfiguration.TermsOfService,
		PrivacyPolicy:  *businessConfiguration.PrivacyPolicy,
	}, nil
}
