-- reverse: drop "mission_criteria_products" table
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
-- reverse: drop "mission_criteria_outlets" table
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
-- reverse: drop "customer_mission_product_progresses" table
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
CREATE INDEX "idx_customer_mission_product_progresses_customer_missio619c0343" ON "customer_mission_product_progresses" ("customer_mission_criteria_progress_id");
CREATE INDEX "idx_customer_mission_product_progresses_product_id" ON "customer_mission_product_progresses" ("product_id");
-- reverse: drop "customer_mission_outlet_progresses" table
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
CREATE INDEX "idx_customer_mission_outlet_progresses_customer_mission7c3ae2a0" ON "customer_mission_outlet_progresses" ("customer_mission_criteria_progress_id");
CREATE INDEX "idx_customer_mission_outlet_progresses_outlet_id" ON "customer_mission_outlet_progresses" ("outlet_id");
-- reverse: create "payment_method_configurations" table
DROP TABLE "payment_method_configurations";
-- reverse: modify "customer_mission_criteria_progresses" table
ALTER TABLE "customer_mission_criteria_progresses" DROP CONSTRAINT "fk_customer_mission_criteria_progresses_mission_criteria";
-- reverse: modify "mission_criteria" table
ALTER TABLE "mission_criteria" DROP CONSTRAINT "fk_mission_criteria_product", DROP CONSTRAINT "fk_mission_criteria_outlet", DROP COLUMN "outlet_id", DROP COLUMN "product_id", ALTER COLUMN "value" TYPE text, ALTER COLUMN "value" DROP DEFAULT, ADD COLUMN "field" character varying(100) NULL, ADD COLUMN "operator" character varying(20) NOT NULL;
-- reverse: modify "point_transactions" table
ALTER TABLE "point_transactions" DROP COLUMN "order_id";
-- reverse: modify "modifier_options" table
ALTER TABLE "modifier_options" DROP COLUMN "image_url";
-- reverse: modify "modifier_groups" table
ALTER TABLE "modifier_groups" DROP COLUMN "dependency_type";
