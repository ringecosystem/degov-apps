----# user subscribe dao rollback
-- Add back the dropped columns
ALTER TABLE dgv_user_subscribed_dao
ADD COLUMN enable_new_proposal int not null default 1,
ADD COLUMN enable_voting_end_reminder int not null default 0;

-- Drop the newly added column
ALTER TABLE dgv_user_subscribed_dao
DROP COLUMN utime;

----# user subscribe proposal rollback
-- Drop the newly added column
ALTER TABLE dgv_user_subscribed_proposal
DROP COLUMN utime;

----# dgv_notification_record rollback
-- Remove the newly added columns
ALTER TABLE dgv_notification_record
DROP COLUMN proposal_id,
DROP COLUMN vote_id,
DROP COLUMN state,
DROP COLUMN payload,
DROP COLUMN event_id;

-- Add back the dropped columns
ALTER TABLE dgv_notification_record
ADD COLUMN chain_name varchar(255) not null,
ADD COLUMN dao_name varchar(255) not null,
ADD COLUMN target_id varchar(255),
ADD COLUMN status varchar(50) not null;

----# subscribe feature rollback
-- Drop the table created in up.sql
DROP TABLE IF EXISTS dgv_subscribed_feature;

----# dgv_proposal_tracking rollback
-- Drop the newly added column
ALTER TABLE dgv_proposal_tracking
DROP COLUMN offset_tracking_vote;

----# dgv_dao rollback
-- Drop the newly added column
ALTER TABLE dgv_dao
DROP COLUMN offset_tracking_proposal;

-- Add back the dropped column
ALTER TABLE dgv_dao
ADD COLUMN last_tracking_block int default 0;

----# dgv_notification_event
drop table if exists dgv_notification_event;
