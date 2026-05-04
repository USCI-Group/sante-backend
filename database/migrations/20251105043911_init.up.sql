-- modify "business_configurations" table
ALTER TABLE "business_configurations" ADD COLUMN "is_tax_included_in_price" boolean NULL DEFAULT true;
-- modify "mission_criteria" table
ALTER TABLE "mission_criteria" ALTER COLUMN "value" TYPE numeric(10,2) USING value::numeric, ALTER COLUMN "value" SET DEFAULT 0;
-- drop index "idx_mission_reward_grants_customer_mission_attempt_id" from table: "mission_reward_grants"
DROP INDEX "idx_mission_reward_grants_customer_mission_attempt_id";
-- rename a column from "customer_mission_attempt_id" to "customer_mission_id"
ALTER TABLE "mission_reward_grants" RENAME COLUMN "customer_mission_attempt_id" TO "customer_mission_id";
-- create index "idx_mission_reward_grants_customer_mission_id" to table: "mission_reward_grants"
CREATE INDEX "idx_mission_reward_grants_customer_mission_id" ON "mission_reward_grants" ("customer_mission_id");
-- modify "missions" table
ALTER TABLE "missions" ADD COLUMN "frequency" character varying(50) NULL;
-- modify "voucher_outlets" table
ALTER TABLE "voucher_outlets" DROP CONSTRAINT "fk_voucher_outlets_outlet", DROP CONSTRAINT "fk_voucher_outlets_voucher";
-- modify "membership_benefits" table
ALTER TABLE "membership_benefits" ADD COLUMN "point_rule_id" uuid NULL, ADD CONSTRAINT "fk_membership_benefits_point_rule" FOREIGN KEY ("point_rule_id") REFERENCES "point_rules" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "vouchers" table
ALTER TABLE "vouchers" ADD COLUMN "eligible_user_type" character varying(50) NULL, ADD COLUMN "product_id" uuid NULL, ADD CONSTRAINT "fk_vouchers_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create "voucher_eligibility_rules" table
CREATE TABLE "voucher_eligibility_rules" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "voucher_id" uuid NULL,
  "eligibility_rule_type" character varying(255) NULL,
  "outlet_id" uuid NULL,
  "product_id" uuid NULL,
  "product_category_id" uuid NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_voucher_eligibility_rules_outlet" FOREIGN KEY ("outlet_id") REFERENCES "outlets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_voucher_eligibility_rules_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_voucher_eligibility_rules_product_category" FOREIGN KEY ("product_category_id") REFERENCES "product_categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_vouchers_voucher_eligibility_rules" FOREIGN KEY ("voucher_id") REFERENCES "vouchers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- drop "customer_limit_rules" table
DROP TABLE "customer_limit_rules";
-- drop "membership_benefit_links" table
DROP TABLE "membership_benefit_links";
-- drop "membership_upgrade_rule_products" table
DROP TABLE "membership_upgrade_rule_products";
