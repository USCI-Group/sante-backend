-- modify "deliveries" table
ALTER TABLE "deliveries" ALTER COLUMN "is_active" DROP DEFAULT;
-- modify "onboardings" table
ALTER TABLE "onboardings" ALTER COLUMN "is_active" DROP DEFAULT;
-- create "mission_reward_grants" table
CREATE TABLE "mission_reward_grants" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_mission_attempt_id" uuid NULL,
  "mission_reward_id" uuid NULL,
  "granted_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_mission_reward_grants_customer_mission_attempt_id" to table: "mission_reward_grants"
CREATE INDEX "idx_mission_reward_grants_customer_mission_attempt_id" ON "mission_reward_grants" ("customer_mission_attempt_id");
-- create index "idx_mission_reward_grants_mission_reward_id" to table: "mission_reward_grants"
CREATE INDEX "idx_mission_reward_grants_mission_reward_id" ON "mission_reward_grants" ("mission_reward_id");
-- modify "point_rules" table
ALTER TABLE "point_rules" ADD COLUMN "day_of_week" bigint NULL;
-- modify "announcements" table
ALTER TABLE "announcements" ALTER COLUMN "is_active" DROP DEFAULT, ADD COLUMN "title" character varying(255) NULL, ADD COLUMN "description" character varying(500) NULL;
-- create "customer_delivery_addresses" table
CREATE TABLE "customer_delivery_addresses" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_id" uuid NULL,
  "street_line1" character varying(150) NULL,
  "street_line2" character varying(150) NULL,
  "street_line3" character varying(150) NULL,
  "city" character varying(50) NULL,
  "state" character varying(50) NULL,
  "postal_code" character varying(20) NULL,
  "country" character varying(50) NULL,
  "name" character varying(255) NULL,
  "is_default" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_customers_customer_delivery_addresses" FOREIGN KEY ("customer_id") REFERENCES "customers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_customer_delivery_addresses_customer_id" to table: "customer_delivery_addresses"
CREATE INDEX "idx_customer_delivery_addresses_customer_id" ON "customer_delivery_addresses" ("customer_id");
-- create "customer_mission_criteria_progresses" table
CREATE TABLE "customer_mission_criteria_progresses" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_mission_id" uuid NULL,
  "mission_criteria_id" uuid NULL,
  "criteria_type" character varying(50) NULL,
  "current_numeric" numeric(10,2) NULL DEFAULT 0,
  "target_numeric" numeric(10,2) NULL DEFAULT 0,
  "is_satisfied" boolean NULL DEFAULT false,
  "last_event_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_customer_mission_criteria_progresses_customer_mission_id" to table: "customer_mission_criteria_progresses"
CREATE INDEX "idx_customer_mission_criteria_progresses_customer_mission_id" ON "customer_mission_criteria_progresses" ("customer_mission_id");
-- create index "idx_customer_mission_criteria_progresses_mission_criteria_id" to table: "customer_mission_criteria_progresses"
CREATE INDEX "idx_customer_mission_criteria_progresses_mission_criteria_id" ON "customer_mission_criteria_progresses" ("mission_criteria_id");
-- create "customer_mission_outlet_progresses" table
CREATE TABLE "customer_mission_outlet_progresses" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_mission_criteria_progress_id" uuid NULL,
  "outlet_id" uuid NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_customer_mission_criteria_progresses_customer_missionc83d834" FOREIGN KEY ("customer_mission_criteria_progress_id") REFERENCES "customer_mission_criteria_progresses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_customer_mission_outlet_progresses_customer_mission7c3ae2a0" to table: "customer_mission_outlet_progresses"
CREATE INDEX "idx_customer_mission_outlet_progresses_customer_mission7c3ae2a0" ON "customer_mission_outlet_progresses" ("customer_mission_criteria_progress_id");
-- create index "idx_customer_mission_outlet_progresses_outlet_id" to table: "customer_mission_outlet_progresses"
CREATE INDEX "idx_customer_mission_outlet_progresses_outlet_id" ON "customer_mission_outlet_progresses" ("outlet_id");
-- create "customer_mission_product_progresses" table
CREATE TABLE "customer_mission_product_progresses" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_mission_criteria_progress_id" uuid NULL,
  "product_id" uuid NULL,
  "current_count" bigint NULL DEFAULT 0,
  "target_count" bigint NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_customer_mission_criteria_progresses_customer_missiond141e6a" FOREIGN KEY ("customer_mission_criteria_progress_id") REFERENCES "customer_mission_criteria_progresses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_customer_mission_product_progresses_customer_missio619c0343" to table: "customer_mission_product_progresses"
CREATE INDEX "idx_customer_mission_product_progresses_customer_missio619c0343" ON "customer_mission_product_progresses" ("customer_mission_criteria_progress_id");
-- create index "idx_customer_mission_product_progresses_product_id" to table: "customer_mission_product_progresses"
CREATE INDEX "idx_customer_mission_product_progresses_product_id" ON "customer_mission_product_progresses" ("product_id");
-- modify "missions" table
ALTER TABLE "missions" ADD COLUMN "image_url" character varying(255) NULL, ADD COLUMN "terms_and_conditions" text NULL;
-- create "customer_missions" table
CREATE TABLE "customer_missions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "customer_id" uuid NULL,
  "mission_id" uuid NULL,
  "status" character varying(20) NULL,
  "progress" bigint NULL DEFAULT 0,
  "started_at" timestamptz NULL,
  "expires_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "ended_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_customer_missions_mission" FOREIGN KEY ("mission_id") REFERENCES "missions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_attempt_unique" to table: "customer_missions"
CREATE UNIQUE INDEX "idx_attempt_unique" ON "customer_missions" ("customer_id", "mission_id");
-- create index "idx_customer_missions_status" to table: "customer_missions"
CREATE INDEX "idx_customer_missions_status" ON "customer_missions" ("status");
-- create "mission_criteria" table
CREATE TABLE "mission_criteria" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "mission_id" uuid NULL,
  "criteria_type" character varying(50) NOT NULL,
  "operator" character varying(20) NOT NULL,
  "membership_id" uuid NULL,
  "field" character varying(100) NULL,
  "value" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_mission_criteria_membership" FOREIGN KEY ("membership_id") REFERENCES "memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_missions_mission_criteria" FOREIGN KEY ("mission_id") REFERENCES "missions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "mission_criteria_outlets" table
CREATE TABLE "mission_criteria_outlets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "mission_criteria_id" uuid NULL,
  "outlet_id" uuid NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_mission_criteria_mission_criteria_outlets" FOREIGN KEY ("mission_criteria_id") REFERENCES "mission_criteria" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_mission_criteria_outlets_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "mission_criteria_products" table
CREATE TABLE "mission_criteria_products" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "mission_criteria_id" uuid NULL,
  "product_id" uuid NULL,
  "criteria_count" bigint NULL DEFAULT 1,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_mission_criteria_mission_criteria_products" FOREIGN KEY ("mission_criteria_id") REFERENCES "mission_criteria" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_mission_criteria_products_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- modify "order_items" table
ALTER TABLE "order_items" DROP CONSTRAINT "fk_order_items_order", ADD CONSTRAINT "fk_orders_order_items" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "selected_modifier_groups" table
ALTER TABLE "selected_modifier_groups" DROP CONSTRAINT "fk_selected_modifier_groups_order_item", ADD CONSTRAINT "fk_order_items_selected_modifier_groups" FOREIGN KEY ("order_item_id") REFERENCES "order_items" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
