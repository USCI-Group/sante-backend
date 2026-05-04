-- reverse: create index "idx_voucher_outlets_deleted_at" to table: "voucher_outlets"
DROP INDEX "idx_voucher_outlets_deleted_at";
-- reverse: create "voucher_outlets" table
DROP TABLE "voucher_outlets";
-- reverse: create index "idx_transactions_transaction_number" to table: "transactions"
DROP INDEX "idx_transactions_transaction_number";
-- reverse: create index "idx_transactions_mol_transaction_id" to table: "transactions"
DROP INDEX "idx_transactions_mol_transaction_id";
-- reverse: create "transactions" table
DROP TABLE "transactions";
-- reverse: create "stocks" table
DROP TABLE "stocks";
-- reverse: create "stock_requested_items" table
DROP TABLE "stock_requested_items";
-- reverse: create "stock_requests" table
DROP TABLE "stock_requests";
-- reverse: create "stock_reports" table
DROP TABLE "stock_reports";
-- reverse: create index "idx_selected_modifier_groups_order_item_id" to table: "selected_modifier_groups"
DROP INDEX "idx_selected_modifier_groups_order_item_id";
-- reverse: create index "idx_selected_modifier_groups_modifier_options_id" to table: "selected_modifier_groups"
DROP INDEX "idx_selected_modifier_groups_modifier_options_id";
-- reverse: create index "idx_selected_modifier_groups_modifier_group_id" to table: "selected_modifier_groups"
DROP INDEX "idx_selected_modifier_groups_modifier_group_id";
-- reverse: create "selected_modifier_groups" table
DROP TABLE "selected_modifier_groups";
-- reverse: create index "idx_recipe_steps_recipe_id" to table: "recipe_steps"
DROP INDEX "idx_recipe_steps_recipe_id";
-- reverse: create "recipe_steps" table
DROP TABLE "recipe_steps";
-- reverse: create index "idx_recipes_product_id" to table: "recipes"
DROP INDEX "idx_recipes_product_id";
-- reverse: create "recipes" table
DROP TABLE "recipes";
-- reverse: create "product_wastage_reports" table
DROP TABLE "product_wastage_reports";
-- reverse: create "product_wastage_types" table
DROP TABLE "product_wastage_types";
-- reverse: create index "idx_product_modifier_mappings_product_id" to table: "product_modifier_mappings"
DROP INDEX "idx_product_modifier_mappings_product_id";
-- reverse: create index "idx_product_modifier_mappings_modifier_group_id" to table: "product_modifier_mappings"
DROP INDEX "idx_product_modifier_mappings_modifier_group_id";
-- reverse: create "product_modifier_mappings" table
DROP TABLE "product_modifier_mappings";
-- reverse: create index "idx_product_ingredient_mappings_product_id" to table: "product_ingredient_mappings"
DROP INDEX "idx_product_ingredient_mappings_product_id";
-- reverse: create index "idx_product_ingredient_mappings_ingredient_id" to table: "product_ingredient_mappings"
DROP INDEX "idx_product_ingredient_mappings_ingredient_id";
-- reverse: create "product_ingredient_mappings" table
DROP TABLE "product_ingredient_mappings";
-- reverse: create index "idx_product_category_mappings_product_sub_category_id" to table: "product_category_mappings"
DROP INDEX "idx_product_category_mappings_product_sub_category_id";
-- reverse: create index "idx_product_category_mappings_product_id" to table: "product_category_mappings"
DROP INDEX "idx_product_category_mappings_product_id";
-- reverse: create index "idx_product_category_mappings_product_category_id" to table: "product_category_mappings"
DROP INDEX "idx_product_category_mappings_product_category_id";
-- reverse: create "product_category_mappings" table
DROP TABLE "product_category_mappings";
-- reverse: create index "idx_product_sub_categories_business_id" to table: "product_sub_categories"
DROP INDEX "idx_product_sub_categories_business_id";
-- reverse: create "product_sub_categories" table
DROP TABLE "product_sub_categories";
-- reverse: create index "idx_product_categories_business_id" to table: "product_categories"
DROP INDEX "idx_product_categories_business_id";
-- reverse: create "product_categories" table
DROP TABLE "product_categories";
-- reverse: create "point_transactions" table
DROP TABLE "point_transactions";
-- reverse: create "point_redemption_transactions" table
DROP TABLE "point_redemption_transactions";
-- reverse: create "point_redemption_rules" table
DROP TABLE "point_redemption_rules";
-- reverse: create index "idx_permissions_name" to table: "permissions"
DROP INDEX "idx_permissions_name";
-- reverse: create index "idx_permissions_group_role_id" to table: "permissions"
DROP INDEX "idx_permissions_group_role_id";
-- reverse: create "permissions" table
DROP TABLE "permissions";
-- reverse: create index "idx_permission_presets_role_id" to table: "permission_presets"
DROP INDEX "idx_permission_presets_role_id";
-- reverse: create "permission_presets" table
DROP TABLE "permission_presets";
-- reverse: create "outlet_products" table
DROP TABLE "outlet_products";
-- reverse: create index "idx_outlet_modifier_options_modifier_options_id" to table: "outlet_modifier_options"
DROP INDEX "idx_outlet_modifier_options_modifier_options_id";
-- reverse: create "outlet_modifier_options" table
DROP TABLE "outlet_modifier_options";
-- reverse: create "outlet_groups_users" table
DROP TABLE "outlet_groups_users";
-- reverse: create "outlet_groups_outlets" table
DROP TABLE "outlet_groups_outlets";
-- reverse: create "outlet_groups" table
DROP TABLE "outlet_groups";
-- reverse: create "order_items" table
DROP TABLE "order_items";
-- reverse: create index "idx_order_details_order_id" to table: "order_details"
DROP INDEX "idx_order_details_order_id";
-- reverse: create "order_details" table
DROP TABLE "order_details";
-- reverse: create "orders" table
DROP TABLE "orders";
-- reverse: create "onboardings" table
DROP TABLE "onboardings";
-- reverse: create "notifications" table
DROP TABLE "notifications";
-- reverse: create index "idx_modifier_ingredient_mappings_modifier_options_id" to table: "modifier_ingredient_mappings"
DROP INDEX "idx_modifier_ingredient_mappings_modifier_options_id";
-- reverse: create index "idx_modifier_ingredient_mappings_ingredient_id" to table: "modifier_ingredient_mappings"
DROP INDEX "idx_modifier_ingredient_mappings_ingredient_id";
-- reverse: create "modifier_ingredient_mappings" table
DROP TABLE "modifier_ingredient_mappings";
-- reverse: create index "idx_modifier_options_modifier_group_id" to table: "modifier_options"
DROP INDEX "idx_modifier_options_modifier_group_id";
-- reverse: create "modifier_options" table
DROP TABLE "modifier_options";
-- reverse: create index "idx_modifier_groups_business_id" to table: "modifier_groups"
DROP INDEX "idx_modifier_groups_business_id";
-- reverse: create "modifier_groups" table
DROP TABLE "modifier_groups";
-- reverse: create "mission_rewards" table
DROP TABLE "mission_rewards";
-- reverse: create "missions" table
DROP TABLE "missions";
-- reverse: create "membership_upgrade_rules" table
DROP TABLE "membership_upgrade_rules";
-- reverse: create "membership_benefit_links" table
DROP TABLE "membership_benefit_links";
-- reverse: create "membership_benefits" table
DROP TABLE "membership_benefits";
-- reverse: create "ingredients" table
DROP TABLE "ingredients";
-- reverse: create "feedback_questions" table
DROP TABLE "feedback_questions";
-- reverse: create "expenses_outlets" table
DROP TABLE "expenses_outlets";
-- reverse: create "deliveries" table
DROP TABLE "deliveries";
-- reverse: create "customer_vouchers" table
DROP TABLE "customer_vouchers";
-- reverse: create index "idx_vouchers_voucher_code" to table: "vouchers"
DROP INDEX "idx_vouchers_voucher_code";
-- reverse: create "vouchers" table
DROP TABLE "vouchers";
-- reverse: create "discounts" table
DROP TABLE "discounts";
-- reverse: create index "idx_customer_product_purchase_stats_product_id" to table: "customer_product_purchase_stats"
DROP INDEX "idx_customer_product_purchase_stats_product_id";
-- reverse: create "customer_product_purchase_stats" table
DROP TABLE "customer_product_purchase_stats";
-- reverse: create index "idx_customer_membership_stats_customer_id" to table: "customer_membership_stats"
DROP INDEX "idx_customer_membership_stats_customer_id";
-- reverse: create index "idx_customer_membership_stats_business_id" to table: "customer_membership_stats"
DROP INDEX "idx_customer_membership_stats_business_id";
-- reverse: create "customer_membership_stats" table
DROP TABLE "customer_membership_stats";
-- reverse: create "customer_memberships" table
DROP TABLE "customer_memberships";
-- reverse: create "customer_limit_rules" table
DROP TABLE "customer_limit_rules";
-- reverse: create "point_rules" table
DROP TABLE "point_rules";
-- reverse: create "memberships" table
DROP TABLE "memberships";
-- reverse: create "customer_favourite_products" table
DROP TABLE "customer_favourite_products";
-- reverse: create "customers" table
DROP TABLE "customers";
-- reverse: create index "idx_products_modifier_options_id" to table: "products"
DROP INDEX "idx_products_modifier_options_id";
-- reverse: create index "idx_products_business_id" to table: "products"
DROP INDEX "idx_products_business_id";
-- reverse: create "products" table
DROP TABLE "products";
-- reverse: create "business_configurations" table
DROP TABLE "business_configurations";
-- reverse: create "announcements" table
DROP TABLE "announcements";
-- reverse: create "activity_logs" table
DROP TABLE "activity_logs";
-- reverse: create "users" table
DROP TABLE "users";
-- reverse: create "outlets" table
DROP TABLE "outlets";
-- reverse: create index "idx_group_roles_role_id" to table: "group_roles"
DROP INDEX "idx_group_roles_role_id";
-- reverse: create index "idx_group_roles_business_id" to table: "group_roles"
DROP INDEX "idx_group_roles_business_id";
-- reverse: create "group_roles" table
DROP TABLE "group_roles";
-- reverse: create index "idx_roles_business_id" to table: "roles"
DROP INDEX "idx_roles_business_id";
-- reverse: create "roles" table
DROP TABLE "roles";
-- reverse: create index "idx_user_tokens_user_id" to table: "user_tokens"
DROP INDEX "idx_user_tokens_user_id";
-- reverse: create "user_tokens" table
DROP TABLE "user_tokens";
-- reverse: create index "idx_info_type" to table: "system_data"
DROP INDEX "idx_info_type";
-- reverse: create "system_data" table
DROP TABLE "system_data";
-- reverse: create index "idx_product_images_product_id" to table: "product_images"
DROP INDEX "idx_product_images_product_id";
-- reverse: create "product_images" table
DROP TABLE "product_images";
-- reverse: create "caches" table
DROP TABLE "caches";
-- reverse: create "businesses" table
DROP TABLE "businesses";
-- reverse: create "merchant_secrets" table
DROP TABLE "merchant_secrets";
-- reverse: create "app_versions" table
DROP TABLE "app_versions";
-- reverse: create index "idx_membership_upgrade_rule_products_deleted_at" to table: "membership_upgrade_rule_products"
DROP INDEX "idx_membership_upgrade_rule_products_deleted_at";
-- reverse: create "membership_upgrade_rule_products" table
DROP TABLE "membership_upgrade_rule_products";
-- reverse: create "customer_tokens" table
DROP TABLE "customer_tokens";
