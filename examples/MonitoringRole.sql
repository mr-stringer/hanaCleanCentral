CREATE USER HCCADMIN PASSWORD "^b3ol345JjnSDaswj^5j6y";
ALTER USER HCCADMIN DISABLE PASSWORD LIFETIME;
CREATE ROLE HCCROLE;
GRANT HCCROLE TO HCCADMIN;
GRANT TRACE ADMIN to HCCADMIN;
GRANT MONITORING TO HCCROLE;
GRANT SELECT ON _SYS_STATISTICS.STATISTICS_ALERTS_BASE TO HCCROLE;
GRANT LOG ADMIN TO HCCROLE;
