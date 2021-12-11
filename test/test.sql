use drawwwingame;
drop table user;
create table user(
	uuid bigint not null primary key,
	tempid char(20) not null,
    name varchar(50) not null,
	email varchar(256) not null,
	password varchar(100) not null,
    expire_tempid_at datetime not null,
    send_email_count int not null,
    send_last_email_at datetime not null,
    email_authorized bit(1) not null,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
insert into user(uuid, tempid, name, email,password,
expire_tempid_at,send_email_count,send_last_email_at,email_authorized,group_id)
values(456,"test","test","test email","rtttt",
now(),0,now(),false,-1);
select * from user;
delete from user where name="test77";
delete from user;

select * from secret;


drop table message;
create table message(
	id bigint not null primary key,
    uuid bigint not null,
    name varchar(50) not null,
    group_id int not null,
    type enum("info", "lines","mark:writer", "mark:answer","text:writer", "text:answer") not null,
    info varchar(100) not null,
    message varchar(1000) not null,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
drop table message_mark;
create table message_mark(
	uuid bigint not null,
    group_id int not null,
    message_id int not null,
    mark char(1) not null,
    type varchar(20) not null,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
drop table user_group;
create table user_group(
	uuid bigint not null primary key,
    group_id int not null,
    admin bit(1) not null,
    can_answer bit(1) not null,
    can_writer bit(1) not null,
    ready bit(1) not null,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);


select * from message;
select * from message_mark;
delete from user;
delete from message_mark;
delete from message;
delete from user_group;
select * from user;
SELECT COUNT(uuid) FROM message_mark WHERE message_id=0 AND mark="A";
SELECT COUNT(uuid) FROM message_mark WHERE message_id=1 AND mark="A" UNION SELECT COUNT(uuid) FROM message_mark WHERE message_id=1 AND mark="B";