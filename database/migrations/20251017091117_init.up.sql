-- modify "modifier_groups" table
ALTER TABLE "modifier_groups" ADD COLUMN "dependency_type" character varying(50) NULL;
-- modify "modifier_options" table
ALTER TABLE "modifier_options" ADD COLUMN "image_url" character varying(255) NULL;
-- modify "point_transactions" table
ALTER TABLE "point_transactions" ADD COLUMN "order_id" uuid NULL;
-- modify "mission_criteria" table
ALTER TABLE "mission_criteria" DROP COLUMN "operator", DROP COLUMN "field", ADD COLUMN "product_id" uuid NULL, ADD COLUMN "outlet_id" uuid NULL, ADD CONSTRAINT "fk_mission_criteria_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "fk_mission_criteria_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "customer_mission_criteria_progresses" table
ALTER TABLE "customer_mission_criteria_progresses" ADD CONSTRAINT "fk_customer_mission_criteria_progresses_mission_criteria" FOREIGN KEY ("mission_criteria_id") REFERENCES "mission_criteria" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create "payment_method_configurations" table
CREATE TABLE "payment_method_configurations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "business_id" uuid NULL,
  "payment_method" character varying(255) NULL,
  "payment_channel" character varying(255) NULL,
  "payment_platform" character varying(255) NULL,
  "payment_channel_code" character varying(50) NULL,
  "is_active" boolean NULL DEFAULT true,
  "is_maintenance" boolean NULL DEFAULT false,
  "is_visible" boolean NULL DEFAULT true,
  "min_amount" numeric(10,2) NULL,
  "max_amount" numeric(10,2) NULL,
  "processing_fee" numeric(10,2) NULL,
  "processing_fee_rate" numeric(5,4) NULL,
  "currency" character varying(3) NULL DEFAULT 'MYR',
  "settlement_days" integer NULL,
  "priority" integer NULL DEFAULT 0,
  "display_name" character varying(255) NULL,
  "description" text NULL,
  "icon_url" character varying(500) NULL,
  "color_code" character varying(7) NULL,
  "compliance_status" character varying(50) NULL DEFAULT 'active',
  "compliance_notes" text NULL,
  "valid_from" timestamptz NULL,
  "valid_until" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_payment_method_configurations_business" FOREIGN KEY ("business_id") REFERENCES "businesses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- drop "customer_mission_outlet_progresses" table
DROP TABLE "customer_mission_outlet_progresses";
-- drop "customer_mission_product_progresses" table
DROP TABLE "customer_mission_product_progresses";
-- drop "mission_criteria_outlets" table
DROP TABLE "mission_criteria_outlets";
-- drop "mission_criteria_products" table
DROP TABLE "mission_criteria_products";
