-- create "notification_queues" table
CREATE TABLE "notification_queues" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "queue_type" varchar(100) NULL,
  "queue_status" varchar(100) NULL,
  "title" varchar(255) NULL,
  "body" text NULL,
  "action_url" text NULL,
  "image_url" text NULL,
  "notification_type" text NULL,
  "order_id" uuid NULL,
  "plan_to_send_at" timestamptz NULL,
  "send_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "failed_at" timestamptz NULL,
  "cancelled_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);

CREATE INDEX "idx_queue_type" ON "notification_queues" ("queue_type");
CREATE INDEX "idx_queue_status" ON "notification_queues" ("queue_status");
CREATE INDEX "idx_notification_type" ON "notification_queues" ("notification_type");
CREATE INDEX "idx_order_id" ON "notification_queues" ("order_id");
CREATE INDEX "idx_plan_to_send_at" ON "notification_queues" ("plan_to_send_at");
CREATE INDEX "idx_send_at" ON "notification_queues" ("send_at");
CREATE INDEX "idx_created_at" ON "notification_queues" ("created_at");
CREATE INDEX "idx_deleted_at" ON "notification_queues" ("deleted_at");
