-- modify "business_configurations" table
ALTER TABLE "business_configurations" ADD COLUMN "terms_of_service" text NULL;
-- modify "membership_upgrade_rules" table
ALTER TABLE "membership_upgrade_rules" ADD COLUMN "product_id" uuid NULL, ADD CONSTRAINT "fk_membership_upgrade_rules_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
