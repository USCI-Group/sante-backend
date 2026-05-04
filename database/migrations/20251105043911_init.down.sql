-- reverse: drop "membership_upgrade_rule_products" table
CREATE TABLE "membership_upgrade_rule_products" (
  "membership_upgrade_rule_id" uuid NOT NULL,
  "product_id" uuid NOT NULL,
  "quantity_required" bigint NULL DEFAULT 1,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("membership_upgrade_rule_id", "product_id")
);
CREATE INDEX "idx_membership_upgrade_rule_products_deleted_at" ON "membership_upgrade_rule_products" ("deleted_at");
-- reverse: drop "membership_benefit_links" table
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
-- reverse: drop "customer_limit_rules" table
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
-- reverse: create "voucher_eligibility_rules" table
DROP TABLE "voucher_eligibility_rules";
-- reverse: modify "vouchers" table
ALTER TABLE "vouchers" DROP CONSTRAINT "fk_vouchers_product", DROP COLUMN "product_id", DROP COLUMN "eligible_user_type";
-- reverse: modify "membership_benefits" table
ALTER TABLE "membership_benefits" DROP CONSTRAINT "fk_membership_benefits_point_rule", DROP COLUMN "point_rule_id";
-- reverse: modify "voucher_outlets" table
ALTER TABLE "voucher_outlets" ADD CONSTRAINT "fk_voucher_outlets_voucher" FOREIGN KEY ("voucher_id") REFERENCES "vouchers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "fk_voucher_outlets_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "missions" table
ALTER TABLE "missions" DROP COLUMN "frequency";
-- reverse: create index "idx_mission_reward_grants_customer_mission_id" to table: "mission_reward_grants"
DROP INDEX "idx_mission_reward_grants_customer_mission_id";
-- reverse: rename a column from "customer_mission_attempt_id" to "customer_mission_id"
ALTER TABLE "mission_reward_grants" RENAME COLUMN "customer_mission_id" TO "customer_mission_attempt_id";
-- reverse: drop index "idx_mission_reward_grants_customer_mission_attempt_id" from table: "mission_reward_grants"
CREATE INDEX "idx_mission_reward_grants_customer_mission_attempt_id" ON "mission_reward_grants" ("customer_mission_attempt_id");
-- reverse: modify "mission_criteria" table
ALTER TABLE "mission_criteria" ALTER COLUMN "value" TYPE text, ALTER COLUMN "value" DROP DEFAULT;
-- reverse: modify "business_configurations" table
ALTER TABLE "business_configurations" DROP COLUMN "is_tax_included_in_price";
