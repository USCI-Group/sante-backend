-- reverse: modify "membership_upgrade_rules" table
ALTER TABLE "membership_upgrade_rules" DROP CONSTRAINT "fk_membership_upgrade_rules_product", DROP COLUMN "product_id";
-- reverse: modify "business_configurations" table
ALTER TABLE "business_configurations" DROP COLUMN "terms_of_service";
