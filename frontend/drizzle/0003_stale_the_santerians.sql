ALTER TABLE `session` ADD `impersonated_by` text;--> statement-breakpoint
ALTER TABLE `user` ADD `role` text;--> statement-breakpoint
ALTER TABLE `user` ADD `banned` integer DEFAULT false;--> statement-breakpoint
ALTER TABLE `user` ADD `ban_reason` text;--> statement-breakpoint
ALTER TABLE `user` ADD `ban_expires` integer;--> statement-breakpoint
UPDATE `user` SET `role` = 'user' WHERE `role` IS NULL;--> statement-breakpoint
UPDATE `user` SET `role` = 'admin'
WHERE `id` = (SELECT `id` FROM `user` ORDER BY `created_at` ASC LIMIT 1)
  AND NOT EXISTS (SELECT 1 FROM `user` WHERE `role` = 'admin');