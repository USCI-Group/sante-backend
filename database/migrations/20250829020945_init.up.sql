-- create "customer_tokens" table
CREATE TABLE "customer_tokens" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_id" uuid NULL,
  "refresh_token" character varying(255) NULL,
  "expired_at" timestamptz NULL,
  "device_id" character varying(255) NULL,
  "fcm_token" character varying(255) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create "membership_upgrade_rule_products" table
CREATE TABLE "membership_upgrade_rule_products" (
  "membership_upgrade_rule_id" uuid NOT NULL,
  "product_id" uuid NOT NULL,
  "quantity_required" bigint NULL DEFAULT 1,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("membership_upgrade_rule_id", "product_id")
);
-- create index "idx_membership_upgrade_rule_products_deleted_at" to table: "membership_upgrade_rule_products"
CREATE INDEX "idx_membership_upgrade_rule_products_deleted_at" ON "membership_upgrade_rule_products" ("deleted_at");
-- create "app_versions" table
CREATE TABLE "app_versions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "app_package_name" character varying(255) NULL,
  "platform" character varying(100) NULL,
  "version_name" character varying(255) NULL,
  "version_code" character varying(255) NULL,
  "minimum_version_name" character varying(255) NULL,
  "minimum_version_code" character varying(255) NULL,
  "release_note" text NULL,
  "download_url" character varying(255) NULL,
  "mandatory_update" boolean NULL,
  "release_date" timestamptz NULL,
  "environment" character varying(100) NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create "merchant_secrets" table
CREATE TABLE "merchant_secrets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "outlet_id" uuid NULL,
  "fiuu_application_code" character varying(255) NULL,
  "fiuu_secret_key" character varying(255) NULL,
  "e_invoice_auth_token" character varying(2048) NULL,
  "e_invoice_auth_token_expiry" timestamptz NULL,
  "grab_store_id" character varying(255) NULL,
  "grab_integration_status" character varying(255) NULL,
  "grab_menu_sync_state" character varying(50) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_merchant_secrets_outlet_id" UNIQUE ("outlet_id")
);
-- create "businesses" table
CREATE TABLE "businesses" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying(255) NULL,
  "email" character varying(255) NULL,
  "phone" character varying(255) NULL,
  "street_line1" character varying(150) NULL,
  "street_line2" character varying(150) NULL,
  "street_line3" character varying(150) NULL,
  "city" character varying(50) NULL,
  "state" character varying(50) NULL,
  "postal_code" character varying(20) NULL,
  "country" character varying(50) NULL,
  "website" character varying(255) NULL,
  "registration_number" character varying(255) NULL,
  "logo_url" character varying(255) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create "caches" table
CREATE TABLE "caches" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "key" character varying(255) NULL,
  "value" text NULL,
  "expiry" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create "product_images" table
CREATE TABLE "product_images" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "product_id" uuid NULL,
  "image_url" character varying(255) NULL,
  "image_format" character varying(255) NULL,
  "image_size" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_product_images_product_id" to table: "product_images"
CREATE INDEX "idx_product_images_product_id" ON "product_images" ("product_id");
-- create "system_data" table
CREATE TABLE "system_data" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "info_type" character varying(255) NULL,
  "info_value" character varying(255) NULL,
  "expiry" timestamptz NULL,
  "is_encrypted" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_info_type" to table: "system_data"
CREATE UNIQUE INDEX "idx_info_type" ON "system_data" ("info_type");
-- create "user_tokens" table
CREATE TABLE "user_tokens" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NULL,
  "refresh_token" character varying(255) NULL,
  "device_id" character varying(255) NULL,
  "expired_at" timestamptz NULL,
  "is_staff_login" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_user_tokens_user_id" to table: "user_tokens"
CREATE INDEX "idx_user_tokens_user_id" ON "user_tokens" ("user_id");
-- create "roles" table
CREATE TABLE "roles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying(255) NULL,
  "description" character varying(255) NULL,
  "role_type" character varying(255) NULL,
  "business_id" uuid NULL,
  "has_outlet_group" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_roles_business_id" to table: "roles"
CREATE INDEX "idx_roles_business_id" ON "roles" ("business_id");
-- create "group_roles" table
CREATE TABLE "group_roles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "role_id" uuid NULL,
  "business_id" uuid NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_group_roles_role" FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_group_roles_business_id" to table: "group_roles"
CREATE INDEX "idx_group_roles_business_id" ON "group_roles" ("business_id");
-- create index "idx_group_roles_role_id" to table: "group_roles"
CREATE INDEX "idx_group_roles_role_id" ON "group_roles" ("role_id");
-- create "outlets" table
CREATE TABLE "outlets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "name" character varying(255) NULL,
  "email" character varying(255) NULL,
  "phone" character varying(255) NULL,
  "street_line1" character varying(150) NULL,
  "street_line2" character varying(150) NULL,
  "street_line3" character varying(150) NULL,
  "city" character varying(50) NULL,
  "state" character varying(50) NULL,
  "postal_code" character varying(20) NULL,
  "country" character varying(50) NULL,
  "website" character varying(255) NULL,
  "id_type" character varying(255) NULL,
  "registration_number" character varying(255) NULL,
  "tin" character varying(255) NULL,
  "image_url" character varying(255) NULL,
  "outlet_status" character varying(10) NULL DEFAULT 'closed',
  "outlet_static_qr" character varying(255) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_outlets_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "outlet_id" uuid NULL,
  "first_name" character varying(255) NULL,
  "surname" character varying(255) NULL,
  "group_role_id" uuid NULL,
  "identification_no" character varying(255) NULL,
  "employee_no" character varying(255) NULL,
  "email" character varying(255) NULL,
  "pwd" character varying(255) NULL,
  "phone" character varying(255) NULL,
  "street_line1" character varying(150) NULL,
  "street_line2" character varying(150) NULL,
  "street_line3" character varying(150) NULL,
  "city" character varying(50) NULL,
  "state" character varying(50) NULL,
  "postal_code" character varying(20) NULL,
  "country" character varying(50) NULL,
  "status" character varying(255) NULL DEFAULT 'ACTIVE',
  "fcm_device_token" character varying(255) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_users_group_role" FOREIGN KEY ("group_role_id") REFERENCES "group_roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_users_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "activity_logs" table
CREATE TABLE "activity_logs" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "activity" character varying(255) NULL,
  "status" character varying(255) NULL,
  "action_by_user_id" uuid NULL,
  "action_by" character varying(255) NULL,
  "details" text NULL,
  "error_message" character varying(255) NULL DEFAULT NULL::character varying,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_activity_logs_user" FOREIGN KEY ("action_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "announcements" table
CREATE TABLE "announcements" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "image_url" character varying(500) NULL,
  "is_active" boolean NULL DEFAULT true,
  "sort_order" bigint NULL,
  "start_date" timestamptz NULL,
  "end_date" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_announcements_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "business_configurations" table
CREATE TABLE "business_configurations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "grab_client_id" character varying(255) NULL,
  "grab_client_secret" character varying(255) NULL,
  "service_charge_percentage" numeric(5,2) NULL DEFAULT 0,
  "service_tax_percentage" numeric(5,2) NULL DEFAULT 0,
  "is_logout_button_visible" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_business_configurations_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "products" table
CREATE TABLE "products" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "name" character varying(255) NULL,
  "description" text NULL,
  "cost" numeric NULL,
  "base_price" numeric NULL,
  "price" numeric NULL,
  "is_standalone_purchase" boolean NULL DEFAULT true,
  "is_addon" boolean NULL DEFAULT false,
  "image_url" character varying(255) NULL,
  "is_active" boolean NULL DEFAULT true,
  "is_store_outlet" boolean NULL DEFAULT true,
  "is_grab_food" boolean NULL DEFAULT false,
  "is_shopee_food" boolean NULL DEFAULT false,
  "sort_order" integer NULL DEFAULT 0,
  "modifier_options_id" uuid NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_products_business_id" to table: "products"
CREATE INDEX "idx_products_business_id" ON "products" ("business_id");
-- create index "idx_products_modifier_options_id" to table: "products"
CREATE INDEX "idx_products_modifier_options_id" ON "products" ("modifier_options_id");
-- create "customers" table
CREATE TABLE "customers" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "first_name" character varying(255) NULL,
  "last_name" character varying(255) NULL,
  "email" character varying(255) NULL,
  "password" character varying(255) NULL,
  "phone_number" character varying(255) NULL,
  "date_of_birth" date NULL,
  "email_verified" boolean NULL DEFAULT false,
  "is_newsletter_subscribed" boolean NULL DEFAULT false,
  "is_agree_to_terms" boolean NULL DEFAULT false,
  "is_agree_to_privacy_policy" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create "customer_favourite_products" table
CREATE TABLE "customer_favourite_products" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_id" uuid NULL,
  "product_id" uuid NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_customer_favourite_products_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_customers_customer_favourite_products" FOREIGN KEY ("customer_id") REFERENCES "customers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "memberships" table
CREATE TABLE "memberships" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "tier_name" character varying(255) NULL,
  "tier_level" bigint NULL DEFAULT 0,
  "tier_image" character varying(255) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create "point_rules" table
CREATE TABLE "point_rules" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "membership_id" uuid NULL,
  "business_id" uuid NULL,
  "name" character varying(255) NULL,
  "action_type" character varying(255) NULL,
  "points_earned" bigint NULL,
  "points_multiplier" bigint NULL,
  "min_amount" numeric NULL,
  "max_points_per_action" bigint NULL,
  "max_points_per_day" bigint NULL,
  "max_points_per_month" bigint NULL,
  "max_points_per_year" bigint NULL,
  "is_active" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_point_rules_membership" FOREIGN KEY ("membership_id") REFERENCES "memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "customer_limit_rules" table
CREATE TABLE "customer_limit_rules" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "point_rule_id" uuid NULL,
  "max_points_per_customer" bigint NULL,
  "max_points_per_customer_per_day" bigint NULL,
  "max_points_per_customer_per_month" bigint NULL,
  "max_points_per_customer_per_year" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_customer_limit_rules_point_rule" FOREIGN KEY ("point_rule_id") REFERENCES "point_rules" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "customer_memberships" table
CREATE TABLE "customer_memberships" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_id" uuid NULL,
  "membership_id" uuid NULL,
  "expiry_date" date NULL,
  "points" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_customer_memberships_customer" FOREIGN KEY ("customer_id") REFERENCES "customers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_customer_memberships_membership" FOREIGN KEY ("membership_id") REFERENCES "memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "customer_membership_stats" table
CREATE TABLE "customer_membership_stats" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_id" uuid NULL,
  "business_id" uuid NULL,
  "total_spending_amount" numeric(10,2) NULL DEFAULT 0,
  "monthly_spending_amount" numeric(10,2) NULL DEFAULT 0,
  "quarterly_spending_amount" numeric(10,2) NULL DEFAULT 0,
  "half_yearly_spending_amount" numeric(10,2) NULL DEFAULT 0,
  "yearly_spending_amount" numeric(10,2) NULL DEFAULT 0,
  "total_orders_count" bigint NULL DEFAULT 0,
  "monthly_orders_count" bigint NULL DEFAULT 0,
  "quarterly_orders_count" bigint NULL DEFAULT 0,
  "half_yearly_orders_count" bigint NULL DEFAULT 0,
  "yearly_orders_count" bigint NULL DEFAULT 0,
  "last_review_date" timestamptz NULL,
  "next_review_date" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_customer_membership_stats_business_id" to table: "customer_membership_stats"
CREATE INDEX "idx_customer_membership_stats_business_id" ON "customer_membership_stats" ("business_id");
-- create index "idx_customer_membership_stats_customer_id" to table: "customer_membership_stats"
CREATE INDEX "idx_customer_membership_stats_customer_id" ON "customer_membership_stats" ("customer_id");
-- create "customer_product_purchase_stats" table
CREATE TABLE "customer_product_purchase_stats" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_membership_stats_id" uuid NULL,
  "product_id" uuid NULL,
  "total_quantity" bigint NULL DEFAULT 0,
  "monthly_quantity" bigint NULL DEFAULT 0,
  "quarterly_quantity" bigint NULL DEFAULT 0,
  "yearly_quantity" bigint NULL DEFAULT 0,
  "half_yearly_quantity" bigint NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_customer_membership_stats_product_purchase_stats" FOREIGN KEY ("customer_membership_stats_id") REFERENCES "customer_membership_stats" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_customer_product_purchase_stats_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_customer_product_purchase_stats_product_id" to table: "customer_product_purchase_stats"
CREATE INDEX "idx_customer_product_purchase_stats_product_id" ON "customer_product_purchase_stats" ("product_id");
-- create "discounts" table
CREATE TABLE "discounts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "membership_id" uuid NULL,
  "name" character varying(255) NULL,
  "description" text NULL,
  "discount_type" character varying(50) NULL,
  "value" numeric(10,2) NULL,
  "is_stackable" boolean NULL,
  "usage_type" character varying(50) NULL,
  "max_discount_value" numeric NULL,
  "valid_from" timestamptz NULL,
  "valid_to" timestamptz NULL,
  "is_active" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_discounts_membership" FOREIGN KEY ("membership_id") REFERENCES "memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "vouchers" table
CREATE TABLE "vouchers" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "membership_id" uuid NULL,
  "business_id" uuid NULL,
  "name" character varying(255) NULL,
  "description" text NULL,
  "terms_and_conditions" text NULL,
  "voucher_code" character varying(255) NOT NULL,
  "voucher_image_url" text NULL,
  "voucher_for" character varying(255) NULL,
  "voucher_type" character varying(255) NULL,
  "min_purchase" numeric NULL,
  "max_purchase" numeric NULL,
  "max_redemption" bigint NULL,
  "max_redemption_per_customer" bigint NULL,
  "current_redemptions" bigint NULL DEFAULT 0,
  "current_usage" bigint NULL DEFAULT 0,
  "redeem_value" numeric(10,2) NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "discount_id" uuid NULL,
  "discount_type" character varying(50) NULL,
  "discount_value" numeric(10,2) NULL,
  "is_eligible_for_ranking_climb" boolean NULL DEFAULT false,
  "eligible_order_method" character varying(50) NULL,
  "eligible_platform" character varying(50) NULL,
  "is_stackable" boolean NULL DEFAULT false,
  "is_exclusive" boolean NULL DEFAULT true,
  "is_one_time_use" boolean NULL DEFAULT true,
  "is_mobile_app_only" boolean NULL DEFAULT false,
  "priority" bigint NULL DEFAULT 0,
  "valid_from" timestamptz NULL,
  "valid_to" timestamptz NULL,
  "validity" bigint NULL,
  "created_by" uuid NULL,
  "updated_by" uuid NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_vouchers_discount" FOREIGN KEY ("discount_id") REFERENCES "discounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_vouchers_membership" FOREIGN KEY ("membership_id") REFERENCES "memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_vouchers_voucher_code" to table: "vouchers"
CREATE UNIQUE INDEX "idx_vouchers_voucher_code" ON "vouchers" ("voucher_code");
-- create "customer_vouchers" table
CREATE TABLE "customer_vouchers" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_id" uuid NULL,
  "voucher_id" uuid NULL,
  "voucher_code" character varying(255) NULL,
  "is_redeemed" boolean NULL DEFAULT false,
  "redeemed_at" timestamptz NULL,
  "validity" bigint NULL,
  "valid_from" timestamptz NULL,
  "valid_to" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_customer_vouchers_voucher" FOREIGN KEY ("voucher_id") REFERENCES "vouchers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "deliveries" table
CREATE TABLE "deliveries" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "image_url" character varying(500) NULL,
  "delivery_type" character varying(255) NULL,
  "is_active" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_deliveries_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "expenses_outlets" table
CREATE TABLE "expenses_outlets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "outlet_id" uuid NULL,
  "expenses_category" character varying(255) NULL,
  "expenses_date" date NULL,
  "expenses_amount" numeric(20,2) NULL,
  "expenses_description" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "expenses_attachment_url" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_expenses_outlets_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "feedback_questions" table
CREATE TABLE "feedback_questions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "question" character varying(500) NULL,
  "section" character varying(255) NULL,
  "image_url" character varying(500) NULL,
  "is_active" boolean NULL DEFAULT true,
  "sort_order" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_feedback_questions_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "ingredients" table
CREATE TABLE "ingredients" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "name" character varying(255) NULL,
  "description" text NULL,
  "unit" character varying(255) NULL,
  "quantity" numeric NULL,
  "price_per_unit" numeric NULL,
  "image_url" character varying(255) NULL,
  "sort_order" integer NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ingredients_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "membership_benefits" table
CREATE TABLE "membership_benefits" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "membership_id" uuid NULL,
  "benefit_name" character varying(255) NULL,
  "benefit_description" text NULL,
  "benefit_type" character varying(255) NULL,
  "benefit_value" character varying(255) NULL,
  "benefit_image" character varying(255) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_memberships_benefits" FOREIGN KEY ("membership_id") REFERENCES "memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "membership_benefit_links" table
CREATE TABLE "membership_benefit_links" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "benefit_id" uuid NULL,
  "linked_model_type" character varying(50) NULL,
  "linked_model_id" uuid NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_membership_benefits_links" FOREIGN KEY ("benefit_id") REFERENCES "membership_benefits" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "membership_upgrade_rules" table
CREATE TABLE "membership_upgrade_rules" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "membership_id" uuid NULL,
  "method" character varying(255) NULL,
  "rule_name" character varying(255) NULL,
  "rule_value" numeric(10,2) NULL,
  "review_period" character varying(50) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_memberships_upgrade_rules" FOREIGN KEY ("membership_id") REFERENCES "memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "missions" table
CREATE TABLE "missions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "name" character varying(255) NULL,
  "description" text NULL,
  "cost" bigint NULL,
  "is_active" boolean NULL DEFAULT false,
  "valid_from" timestamptz NULL,
  "valid_to" timestamptz NULL,
  "validity" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create "mission_rewards" table
CREATE TABLE "mission_rewards" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "mission_id" uuid NULL,
  "reward_type" character varying(50) NULL,
  "point_rule_id" uuid NULL,
  "voucher_id" uuid NULL,
  "quantity" bigint NULL DEFAULT 1,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_mission_rewards_point_rule" FOREIGN KEY ("point_rule_id") REFERENCES "point_rules" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_mission_rewards_voucher" FOREIGN KEY ("voucher_id") REFERENCES "vouchers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_missions_mission_rewards" FOREIGN KEY ("mission_id") REFERENCES "missions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_mission_rewards_reward_type" CHECK ((reward_type)::text = ANY ((ARRAY['points'::character varying, 'voucher'::character varying])::text[]))
);
-- create "modifier_groups" table
CREATE TABLE "modifier_groups" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "name" character varying(255) NULL,
  "input_type" character varying(255) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_modifier_groups_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_modifier_groups_business_id" to table: "modifier_groups"
CREATE INDEX "idx_modifier_groups_business_id" ON "modifier_groups" ("business_id");
-- create "modifier_options" table
CREATE TABLE "modifier_options" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "modifier_group_id" uuid NULL,
  "name" character varying(255) NULL,
  "price_adjustment" numeric NULL,
  "sort_order" integer NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_modifier_options_modifier_group" FOREIGN KEY ("modifier_group_id") REFERENCES "modifier_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_modifier_options_modifier_group_id" to table: "modifier_options"
CREATE INDEX "idx_modifier_options_modifier_group_id" ON "modifier_options" ("modifier_group_id");
-- create "modifier_ingredient_mappings" table
CREATE TABLE "modifier_ingredient_mappings" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "ingredient_id" uuid NULL,
  "modifier_options_id" uuid NULL,
  "unit" character varying(255) NULL,
  "quantity" numeric NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_modifier_ingredient_mappings_ingredient" FOREIGN KEY ("ingredient_id") REFERENCES "ingredients" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_modifier_options_ingredient_mappings" FOREIGN KEY ("modifier_options_id") REFERENCES "modifier_options" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_modifier_ingredient_mappings_ingredient_id" to table: "modifier_ingredient_mappings"
CREATE INDEX "idx_modifier_ingredient_mappings_ingredient_id" ON "modifier_ingredient_mappings" ("ingredient_id");
-- create index "idx_modifier_ingredient_mappings_modifier_options_id" to table: "modifier_ingredient_mappings"
CREATE INDEX "idx_modifier_ingredient_mappings_modifier_options_id" ON "modifier_ingredient_mappings" ("modifier_options_id");
-- create "notifications" table
CREATE TABLE "notifications" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NULL,
  "outlet_id" uuid NULL,
  "fcm_device_token" character varying(255) NULL,
  "title" character varying(255) NULL,
  "body" text NULL,
  "data" jsonb NULL,
  "notification_type" text NULL,
  "is_read" boolean NULL DEFAULT false,
  "action_url" text NULL,
  "image_url" text NULL,
  "expired_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_notifications_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_notifications_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "onboardings" table
CREATE TABLE "onboardings" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "title" character varying(255) NULL,
  "description" character varying(255) NULL,
  "image_url" character varying(500) NULL,
  "is_active" boolean NULL DEFAULT true,
  "sort_order" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_onboardings_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "orders" table
CREATE TABLE "orders" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "outlet_id" uuid NULL,
  "user_id" uuid NULL,
  "customer_id" uuid NULL,
  "order_number" character varying(255) NULL,
  "order_date" timestamptz NULL,
  "order_type" character varying(255) NULL,
  "invoice_number" character varying(255) NULL,
  "invoice_date" timestamptz NULL,
  "order_status" character varying(255) NULL,
  "platform" character varying(255) NULL,
  "platform_order_id" character varying(255) NULL,
  "platform_state" character varying(255) NULL,
  "gross_total" numeric(10,2) NULL,
  "net_total" numeric(10,2) NULL,
  "rounded_amount" numeric(10,2) NULL,
  "rounded_net_total" numeric(10,2) NULL,
  "amount_received" numeric(10,2) NULL,
  "tax_charge" numeric(10,2) NULL,
  "tax_percentage" numeric(10,2) NULL,
  "service_charge" numeric(10,2) NULL,
  "service_charge_percentage" numeric(10,2) NULL,
  "discount_type" character varying(255) NULL,
  "discount_amount" numeric(10,2) NULL,
  "discount_percentage" numeric(10,2) NULL,
  "payment_method" character varying(255) NULL,
  "payment_status" character varying(255) NULL,
  "notes" text NULL,
  "table_number" character varying(255) NULL,
  "e_invoice_submission_id" character varying(255) NULL,
  "e_invoice_status" character varying(255) NULL,
  "e_invoice_url" character varying(512) NULL,
  "e_invoice_rejected_reason" character varying(255) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_orders_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_orders_customer" FOREIGN KEY ("customer_id") REFERENCES "customers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_orders_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_orders_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "order_details" table
CREATE TABLE "order_details" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NULL,
  "grab_short_order_num" character varying(20) NULL,
  "customer_name" character varying(50) NULL,
  "customer_phone" character varying(50) NULL,
  "customer_address" character varying(255) NULL,
  "customer_latitude" numeric(10,8) NULL,
  "customer_longitude" numeric(10,8) NULL,
  "estimated_order_ready_time" timestamptz NULL,
  "max_order_ready_time" timestamptz NULL,
  "new_order_ready_time" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_orders_order_details" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_order_details_order_id" to table: "order_details"
CREATE INDEX "idx_order_details_order_id" ON "order_details" ("order_id");
-- create "order_items" table
CREATE TABLE "order_items" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NULL,
  "product_id" uuid NULL,
  "quantity" integer NULL,
  "unit_price" numeric(10,2) NULL,
  "sub_total" numeric(10,2) NULL,
  "item_notes" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_order_items_order" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_order_items_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "outlet_groups" table
CREATE TABLE "outlet_groups" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying(255) NULL,
  "description" character varying(255) NULL,
  "business_id" uuid NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_outlet_groups_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "outlet_groups_outlets" table
CREATE TABLE "outlet_groups_outlets" (
  "outlet_group_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "outlet_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  PRIMARY KEY ("outlet_group_id", "outlet_id"),
  CONSTRAINT "fk_outlet_groups_outlets_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_outlet_groups_outlets_outlet_group" FOREIGN KEY ("outlet_group_id") REFERENCES "outlet_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "outlet_groups_users" table
CREATE TABLE "outlet_groups_users" (
  "outlet_group_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  PRIMARY KEY ("outlet_group_id", "user_id"),
  CONSTRAINT "fk_outlet_groups_users_outlet_group" FOREIGN KEY ("outlet_group_id") REFERENCES "outlet_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_outlet_groups_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "outlet_modifier_options" table
CREATE TABLE "outlet_modifier_options" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "outlet_id" uuid NULL,
  "modifier_options_id" uuid NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_outlet_modifier_options_modifier_options" FOREIGN KEY ("modifier_options_id") REFERENCES "modifier_options" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_outlet_modifier_options_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_outlet_modifier_options_modifier_options_id" to table: "outlet_modifier_options"
CREATE INDEX "idx_outlet_modifier_options_modifier_options_id" ON "outlet_modifier_options" ("modifier_options_id");
-- create "outlet_products" table
CREATE TABLE "outlet_products" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "outlet_id" uuid NULL,
  "product_id" uuid NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_outlet_products_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_outlet_products_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "permission_presets" table
CREATE TABLE "permission_presets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying(255) NULL,
  "module" character varying(255) NULL,
  "sub_module" character varying(255) NULL,
  "role_id" uuid NULL,
  "enabled" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_roles_permission_presets" FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_permission_presets_role_id" to table: "permission_presets"
CREATE INDEX "idx_permission_presets_role_id" ON "permission_presets" ("role_id");
-- create "permissions" table
CREATE TABLE "permissions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying(255) NULL,
  "module" character varying(255) NULL,
  "sub_module" character varying(255) NULL,
  "group_role_id" uuid NULL,
  "enabled" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_group_roles_permissions" FOREIGN KEY ("group_role_id") REFERENCES "group_roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_permissions_group_role_id" to table: "permissions"
CREATE INDEX "idx_permissions_group_role_id" ON "permissions" ("group_role_id");
-- create index "idx_permissions_name" to table: "permissions"
CREATE INDEX "idx_permissions_name" ON "permissions" ("name");
-- create "point_redemption_rules" table
CREATE TABLE "point_redemption_rules" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "rule_name" character varying(255) NULL,
  "description" character varying(255) NULL,
  "terms_and_conditions" text NULL,
  "min_amount" numeric NULL,
  "max_amount" numeric NULL,
  "max_redemption" bigint NULL,
  "max_redemption_per_customer" bigint NULL,
  "max_redemption_per_day" bigint NULL,
  "max_redemption_per_month" bigint NULL,
  "max_redemption_per_year" bigint NULL,
  "max_redemption_per_transaction" bigint NULL,
  "max_redemption_per_order" bigint NULL,
  "applicable_tier_level" bigint NULL,
  "is_active" boolean NULL,
  "valid_from" timestamptz NULL,
  "valid_to" timestamptz NULL,
  "is_allow_all_payment_type" boolean NULL,
  "discount_percentage" numeric NULL,
  "exchange_rate" numeric NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_point_redemption_rules_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "point_redemption_transactions" table
CREATE TABLE "point_redemption_transactions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "point_redemption_rule_id" uuid NULL,
  "customer_id" uuid NULL,
  "points_redeemed" bigint NULL,
  "redeemed_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_point_redemption_transactions_customer" FOREIGN KEY ("customer_id") REFERENCES "customers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_point_redemption_transactions_point_redemption_rule" FOREIGN KEY ("point_redemption_rule_id") REFERENCES "point_redemption_rules" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "point_transactions" table
CREATE TABLE "point_transactions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "point_rule_id" uuid NULL,
  "customer_id" uuid NULL,
  "points_earned" bigint NULL,
  "earned_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_point_transactions_customer" FOREIGN KEY ("customer_id") REFERENCES "customers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_point_transactions_point_rule" FOREIGN KEY ("point_rule_id") REFERENCES "point_rules" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "product_categories" table
CREATE TABLE "product_categories" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying(255) NULL,
  "description" text NULL,
  "business_id" uuid NULL,
  "sort_order" integer NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_product_categories_business_id" to table: "product_categories"
CREATE INDEX "idx_product_categories_business_id" ON "product_categories" ("business_id");
-- create "product_sub_categories" table
CREATE TABLE "product_sub_categories" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "name" character varying(255) NULL,
  "description" text NULL,
  "sort_order" integer NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_product_sub_categories_business_id" to table: "product_sub_categories"
CREATE INDEX "idx_product_sub_categories_business_id" ON "product_sub_categories" ("business_id");
-- create "product_category_mappings" table
CREATE TABLE "product_category_mappings" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "product_id" uuid NULL,
  "product_category_id" uuid NULL,
  "product_sub_category_id" uuid NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_product_category_mappings_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_product_category_mappings_product_category" FOREIGN KEY ("product_category_id") REFERENCES "product_categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_product_category_mappings_product_sub_category" FOREIGN KEY ("product_sub_category_id") REFERENCES "product_sub_categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_product_category_mappings_product_category_id" to table: "product_category_mappings"
CREATE INDEX "idx_product_category_mappings_product_category_id" ON "product_category_mappings" ("product_category_id");
-- create index "idx_product_category_mappings_product_id" to table: "product_category_mappings"
CREATE INDEX "idx_product_category_mappings_product_id" ON "product_category_mappings" ("product_id");
-- create index "idx_product_category_mappings_product_sub_category_id" to table: "product_category_mappings"
CREATE INDEX "idx_product_category_mappings_product_sub_category_id" ON "product_category_mappings" ("product_sub_category_id");
-- create "product_ingredient_mappings" table
CREATE TABLE "product_ingredient_mappings" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "product_id" uuid NULL,
  "ingredient_id" uuid NULL,
  "unit" character varying(255) NULL,
  "quantity" numeric NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_product_ingredient_mappings_ingredient" FOREIGN KEY ("ingredient_id") REFERENCES "ingredients" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_product_ingredient_mappings_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_product_ingredient_mappings_ingredient_id" to table: "product_ingredient_mappings"
CREATE INDEX "idx_product_ingredient_mappings_ingredient_id" ON "product_ingredient_mappings" ("ingredient_id");
-- create index "idx_product_ingredient_mappings_product_id" to table: "product_ingredient_mappings"
CREATE INDEX "idx_product_ingredient_mappings_product_id" ON "product_ingredient_mappings" ("product_id");
-- create "product_modifier_mappings" table
CREATE TABLE "product_modifier_mappings" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "product_id" uuid NULL,
  "modifier_group_id" uuid NULL,
  "max_selection" integer NULL DEFAULT 1,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_product_modifier_mappings_modifier_group" FOREIGN KEY ("modifier_group_id") REFERENCES "modifier_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_product_modifier_mappings_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_product_modifier_mappings_modifier_group_id" to table: "product_modifier_mappings"
CREATE INDEX "idx_product_modifier_mappings_modifier_group_id" ON "product_modifier_mappings" ("modifier_group_id");
-- create index "idx_product_modifier_mappings_product_id" to table: "product_modifier_mappings"
CREATE INDEX "idx_product_modifier_mappings_product_id" ON "product_modifier_mappings" ("product_id");
-- create "product_wastage_types" table
CREATE TABLE "product_wastage_types" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "name" character varying(255) NULL,
  "description" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "sort_order" integer NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create "product_wastage_reports" table
CREATE TABLE "product_wastage_reports" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "outlet_id" uuid NULL,
  "product_id" uuid NULL,
  "wastage_type_id" uuid NULL,
  "wastage_amount" numeric(20,6) NULL,
  "report_date" timestamptz NULL,
  "notes" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_product_wastage_reports_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_product_wastage_reports_wastage_type" FOREIGN KEY ("wastage_type_id") REFERENCES "product_wastage_types" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "recipes" table
CREATE TABLE "recipes" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "product_id" uuid NULL,
  "name" character varying(255) NULL,
  "description" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_recipes_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_recipes_product_id" to table: "recipes"
CREATE INDEX "idx_recipes_product_id" ON "recipes" ("product_id");
-- create "recipe_steps" table
CREATE TABLE "recipe_steps" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "recipe_id" uuid NULL,
  "name" character varying(255) NULL,
  "instruction" text NULL,
  "precedence" integer NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_recipe_steps_recipe" FOREIGN KEY ("recipe_id") REFERENCES "recipes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_recipe_steps_recipe_id" to table: "recipe_steps"
CREATE INDEX "idx_recipe_steps_recipe_id" ON "recipe_steps" ("recipe_id");
-- create "selected_modifier_groups" table
CREATE TABLE "selected_modifier_groups" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_item_id" uuid NULL,
  "modifier_group_id" uuid NULL,
  "modifier_options_id" uuid NULL,
  "modifier_option_quantity" integer NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_selected_modifier_groups_modifier_group" FOREIGN KEY ("modifier_group_id") REFERENCES "modifier_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_selected_modifier_groups_modifier_options" FOREIGN KEY ("modifier_options_id") REFERENCES "modifier_options" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_selected_modifier_groups_order_item" FOREIGN KEY ("order_item_id") REFERENCES "order_items" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_selected_modifier_groups_modifier_group_id" to table: "selected_modifier_groups"
CREATE INDEX "idx_selected_modifier_groups_modifier_group_id" ON "selected_modifier_groups" ("modifier_group_id");
-- create index "idx_selected_modifier_groups_modifier_options_id" to table: "selected_modifier_groups"
CREATE INDEX "idx_selected_modifier_groups_modifier_options_id" ON "selected_modifier_groups" ("modifier_options_id");
-- create index "idx_selected_modifier_groups_order_item_id" to table: "selected_modifier_groups"
CREATE INDEX "idx_selected_modifier_groups_order_item_id" ON "selected_modifier_groups" ("order_item_id");
-- create "stock_reports" table
CREATE TABLE "stock_reports" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "outlet_id" uuid NULL,
  "ingredient_id" uuid NULL,
  "sales" numeric NULL,
  "purchases" numeric NULL,
  "transfer_in" numeric NULL,
  "transfer_out" numeric NULL,
  "wastage" numeric NULL,
  "opening" numeric(20,6) NULL,
  "opening_by_system" numeric(20,6) NULL,
  "closing" numeric(20,6) NULL,
  "closing_by_system" numeric(20,6) NULL,
  "variance" numeric(20,6) NULL,
  "cash_opening" numeric(20,2) NULL,
  "cash_closing" numeric(20,2) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_stock_reports_ingredient" FOREIGN KEY ("ingredient_id") REFERENCES "ingredients" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "stock_requests" table
CREATE TABLE "stock_requests" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "requester_outlet_id" uuid NULL,
  "request_date" timestamptz NULL,
  "remarks" text NULL,
  "request_status" character varying(255) NULL,
  "requester_id" uuid NULL,
  "responder_outlet_id" uuid NULL,
  "responder_id" uuid NULL,
  "response_date" timestamptz NULL,
  "response_status" character varying(255) NULL,
  "responder_remarks" text NULL,
  "is_received" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_stock_requests_requester" FOREIGN KEY ("requester_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_stock_requests_requester_outlet" FOREIGN KEY ("requester_outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_stock_requests_responder" FOREIGN KEY ("responder_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_stock_requests_responder_outlet" FOREIGN KEY ("responder_outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "stock_requested_items" table
CREATE TABLE "stock_requested_items" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "stock_request_id" uuid NULL,
  "ingredient_id" uuid NULL,
  "unit_selected" character varying(255) NULL,
  "quantity" numeric(20,6) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_stock_requested_items_ingredient" FOREIGN KEY ("ingredient_id") REFERENCES "ingredients" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_stock_requests_requested_items" FOREIGN KEY ("stock_request_id") REFERENCES "stock_requests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "stocks" table
CREATE TABLE "stocks" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "outlet_id" uuid NULL,
  "ingredient_id" uuid NULL,
  "name" character varying(255) NULL,
  "description" text NULL,
  "small_scale_unit" character varying(255) NULL,
  "large_scale_unit" character varying(255) NULL,
  "small_scale_quantity" numeric NULL,
  "large_scale_quantity" numeric NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_stocks_ingredient" FOREIGN KEY ("ingredient_id") REFERENCES "ingredients" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_stocks_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "transactions" table
CREATE TABLE "transactions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NULL,
  "transaction_number" character varying(255) NULL,
  "mol_transaction_id" character varying(255) NULL,
  "transaction_date" timestamptz NULL,
  "amount" numeric(10,2) NULL,
  "payment_method" character varying(255) NULL,
  "payment_status" character varying(255) NULL,
  "error_code" character varying(255) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_transactions_order" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_transactions_mol_transaction_id" to table: "transactions"
CREATE UNIQUE INDEX "idx_transactions_mol_transaction_id" ON "transactions" ("mol_transaction_id");
-- create index "idx_transactions_transaction_number" to table: "transactions"
CREATE UNIQUE INDEX "idx_transactions_transaction_number" ON "transactions" ("transaction_number");
-- create "voucher_outlets" table
CREATE TABLE "voucher_outlets" (
  "voucher_id" uuid NOT NULL,
  "outlet_id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("voucher_id", "outlet_id"),
  CONSTRAINT "fk_voucher_outlets_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_voucher_outlets_voucher" FOREIGN KEY ("voucher_id") REFERENCES "vouchers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_voucher_outlets_deleted_at" to table: "voucher_outlets"
CREATE INDEX "idx_voucher_outlets_deleted_at" ON "voucher_outlets" ("deleted_at");
