----# user subscribe dao
-- Drop the existing columns
ALTER TABLE dgv_user_subscribed_dao
DROP COLUMN enable_new_proposal,
DROP COLUMN enable_voting_end_reminder;

alter table dgv_user_subscribed_dao
add column utime timestamp;

----# user subscribe proposal
alter table dgv_user_subscribed_proposal
add column utime timestamp;

----# dgv_notification_record
alter table dgv_notification_record
drop column chain_name,
drop column dao_name,
drop column target_id,
drop column status;

alter table dgv_notification_record
add column proposal_id varchar(255) not null,
add column vote_id varchar(255),
add column state varchar(50) not null,
add column payload text,
add column event_id varchar(255) not null;

create unique index uq_notification_record_event_id_user_id on dgv_notification_record (event_id, user_id);

COMMENT ON COLUMN dgv_notification_record.proposal_id IS 'proposal id';

COMMENT ON COLUMN dgv_notification_record.vote_id IS 'vote id';

----# subscribe feature
create table
  if not exists dgv_subscribed_feature (
    id varchar(50) not null,
    chain_id int not null,
    dao_code varchar(255) not null,
    user_id varchar(50) not null,
    user_address varchar(255) not null,
    feature varchar(255) not null,
    strategy varchar(255) not null,
    proposal_id varchar(255),
    ctime timestamp default now (),
    primary key (id)
  );

COMMENT ON COLUMN dgv_subscribed_feature.feature IS 'subscribe feature';

COMMENT ON COLUMN dgv_subscribed_feature.strategy IS 'subscribe strategy';

COMMENT ON COLUMN dgv_subscribed_feature.proposal_id IS 'proposal id';

----# dgv_proposal_tracking
alter table dgv_proposal_tracking
add column offset_tracking_vote int default 0;

----# dgv_dao
alter table dgv_dao
drop column last_tracking_block;

alter table dgv_dao
add column offset_tracking_proposal int default 0;

----# dgv_notification_event
create table
  if not exists dgv_notification_event (
    id varchar(50) not null,
    chain_id int not null,
    dao_code varchar(255) not null,
    type varchar(50) not null,
    proposal_id varchar(255) not null,
    vote_id varchar(255),
    reached int not null default 0,
    state varchar(50) not null,
    time_event timestamp,
    payload text,
    ctime timestamp default now (),
    primary key (id)
  );

COMMENT ON COLUMN dgv_notification_event.event_type IS 'event type';
