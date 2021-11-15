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
    group_id int not null,
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
