-- reverse: modify "selected_modifier_groups" table
ALTER TABLE "selected_modifier_groups" DROP CONSTRAINT "fk_order_items_selected_modifier_groups", ADD CONSTRAINT "fk_selected_modifier_groups_order_item" FOREIGN KEY ("order_item_id") REFERENCES "order_items" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- reverse: modify "order_items" table
ALTER TABLE "order_items" DROP CONSTRAINT "fk_orders_order_items", ADD CONSTRAINT "fk_order_items_order" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- reverse: create "mission_criteria_products" table
DROP TABLE "mission_criteria_products";
-- reverse: create "mission_criteria_outlets" table
DROP TABLE "mission_criteria_outlets";
-- reverse: create "mission_criteria" table
DROP TABLE "mission_criteria";
-- reverse: create index "idx_customer_missions_status" to table: "customer_missions"
DROP INDEX "idx_customer_missions_status";
-- reverse: create index "idx_attempt_unique" to table: "customer_missions"
DROP INDEX "idx_attempt_unique";
-- reverse: create "customer_missions" table
DROP TABLE "customer_missions";
-- reverse: modify "missions" table
ALTER TABLE "missions" DROP COLUMN "terms_and_conditions", DROP COLUMN "image_url";
-- reverse: create index "idx_customer_mission_product_progresses_product_id" to table: "customer_mission_product_progresses"
DROP INDEX "idx_customer_mission_product_progresses_product_id";
-- reverse: create index "idx_customer_mission_product_progresses_customer_missio619c0343" to table: "customer_mission_product_progresses"
DROP INDEX "idx_customer_mission_product_progresses_customer_missio619c0343";
-- reverse: create "customer_mission_product_progresses" table
DROP TABLE "customer_mission_product_progresses";
-- reverse: create index "idx_customer_mission_outlet_progresses_outlet_id" to table: "customer_mission_outlet_progresses"
DROP INDEX "idx_customer_mission_outlet_progresses_outlet_id";
-- reverse: create index "idx_customer_mission_outlet_progresses_customer_mission7c3ae2a0" to table: "customer_mission_outlet_progresses"
DROP INDEX "idx_customer_mission_outlet_progresses_customer_mission7c3ae2a0";
-- reverse: create "customer_mission_outlet_progresses" table
DROP TABLE "customer_mission_outlet_progresses";
-- reverse: create index "idx_customer_mission_criteria_progresses_mission_criteria_id" to table: "customer_mission_criteria_progresses"
DROP INDEX "idx_customer_mission_criteria_progresses_mission_criteria_id";
-- reverse: create index "idx_customer_mission_criteria_progresses_customer_mission_id" to table: "customer_mission_criteria_progresses"
DROP INDEX "idx_customer_mission_criteria_progresses_customer_mission_id";
-- reverse: create "customer_mission_criteria_progresses" table
DROP TABLE "customer_mission_criteria_progresses";
-- reverse: create index "idx_customer_delivery_addresses_customer_id" to table: "customer_delivery_addresses"
DROP INDEX "idx_customer_delivery_addresses_customer_id";
-- reverse: create "customer_delivery_addresses" table
DROP TABLE "customer_delivery_addresses";
-- reverse: modify "announcements" table
ALTER TABLE "announcements" DROP COLUMN "description", DROP COLUMN "title", ALTER COLUMN "is_active" SET DEFAULT true;
-- reverse: modify "point_rules" table
ALTER TABLE "point_rules" DROP COLUMN "day_of_week";
-- reverse: create index "idx_mission_reward_grants_mission_reward_id" to table: "mission_reward_grants"
DROP INDEX "idx_mission_reward_grants_mission_reward_id";
-- reverse: create index "idx_mission_reward_grants_customer_mission_attempt_id" to table: "mission_reward_grants"
DROP INDEX "idx_mission_reward_grants_customer_mission_attempt_id";
-- reverse: create "mission_reward_grants" table
DROP TABLE "mission_reward_grants";
-- reverse: modify "onboardings" table
ALTER TABLE "onboardings" ALTER COLUMN "is_active" SET DEFAULT true;
-- reverse: modify "deliveries" table
ALTER TABLE "deliveries" ALTER COLUMN "is_active" SET DEFAULT false;
