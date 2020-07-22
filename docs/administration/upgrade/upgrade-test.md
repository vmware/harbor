---
title: Test Harbor Upgrade
weight: 50
---

## Prepare data  
1. Add user usera userb userc userd usere, set usera userb as system admin.  
2. Create project projecta projectc as private, create projectb as public.  
3. Add usera as projecta's admin, userc as developer, and userd as guest. Do the same to projectb and projectc.  
4. Login harbor as usera, push an unsigned image into projecta, then push a signed image to projecta. 
5. Login harbor as userc, push an unsigned image into projecta, then push a signed image to projeca.
6. Login harbor as userd, push each image one time.   
7. Repeat 4 5 6 to projectb and projectc.
8. Add one endpoint to harbor.  
9. Add an immediate replication rule to projeca, a schedule rule to projectb, a manual rule to projectc, trigger each rule one time.  
10. Add 5 system label syslabel1 to syslabel5 and tag syslabel1 and syslabel2 to all unsigned image.    
11. In each project add 5 project label projlabela to projlabele, add projlabela projlabelb and projlabelc to signed image. 
12. Trigger one scan all job to scan all images.(For clair enabled instance)  
13. Update project publicly, content trust, severity and scanning settings.
14. Update Harbor email, token expire read only and scan settings.  
15. Update repository info.   
**NOTE**: Create user step is not needed if auth mode is LDAP.  

# Upgrade

## Follow the upgrade guide  
1. Run db migrator image to backup database.
2. Run db migrator image to migrate database.
3. Install new version harbor.

# After upgrade  
  
1. Confirm users are exist and available(No need for VIC and LDAP Mode).  
2. Confirm users have the correct role.  
3. Confirm labels are existing and labeled correct.(No need for VIC)  
4. Confirm notary signature correct.  
5. Confirm endpoint exist.  
6. Confirm replication rule exist and works well.  
7. Confirm project level settings(publicly, content trust, scan) same as before.  
8. Confirm system level settings(email token expire scan) same as before.  
9. Confirm scan result the same as before upgrade.  
10. Confirm access log the same as before upgrade.  
11. Confirm repository info the same as before.  
12. Confirm other image metadata(e.g. author, size) the same as before. 