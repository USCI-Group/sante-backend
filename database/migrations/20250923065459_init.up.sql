-- modify "customers" table
ALTER TABLE "customers" ADD COLUMN "profile_picture" character varying(500) NULL;
-- modify "orders" table
ALTER TABLE "orders" ADD COLUMN "payment_channel" character varying(255) NULL;
-- create "feedbacks" table
CREATE TABLE "feedbacks" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "feedback_question_id" uuid NULL,
  "customer_id" uuid NULL,
  "rating" bigint NULL,
  "comment" character varying(500) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_feedbacks_customer" FOREIGN KEY ("customer_id") REFERENCES "customers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_feedbacks_feedback_question" FOREIGN KEY ("feedback_question_id") REFERENCES "feedback_questions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_feedbacks_rating" CHECK ((rating >= 1) AND (rating <= 5))
);
