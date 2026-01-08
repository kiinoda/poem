+++
title = "There is no Kubernetes"
date = "2025-04-18"
author = "kiiNODA"
draft=false
+++
When you're a beginner in the DevOps world, you believe that learning Linux will get you a job. Nobody tells you what the real world wants from you. It's not complicated but there are many moving parts. Let's talk history. 🧵

It's 2010. There are no Kubernetes, Docker or serverless. All you have is physical servers. And some VPS providers. Oh, and AWS is already there. But people don't make too much of it yet. It's 2010.

So, you're a sysadmin. No DevOps, yet, that will come later. You need to spin up some website. Throw together a server of sorts, add Apache and PHP to the mix. Add MySQL, set up FTP and point a domain name to your server. It's up there. Great! [HowTo](https://www.youtube.com/watch?v=TrLAx27Npns)

The server fails. You learn after a few days. Learn about Pingdom. Now, if the server fails, you get an email. You're covered. [Pingdom](https://www.pingdom.com)

The server fails again. The database crashed. For some reason, your data is corrupted. Damn! You start from scratch, your users are pissed but that's life. You find a script online that you run once a week, saving your database dump into a directory. [MySQL Dump](https://dev.mysql.com/doc/refman/5.7/en/mysqldump-sql-format.html)

You get tired with backing up your data weekly, manually. You learn of shell scripts and add a cron job to run that script and save your database dump in a separate file daily. You're covered. [Cron](https://ostechnix.com/a-beginners-guide-to-cron-jobs/) [Shell scripting](https://www.youtube.com/watch?v=GtovwKDemnI)

The server fails again after a while. The disks failed. You lost everything. Damn, your backups were local, you lost them, as well. Your users are pissed but that's life. You set everything back up from scratch. You start pushing your files remotely, via FTP. You're covered.

The server is hacked. They delete everything. Your remote FTP backups get deleted because you had login info in the backup script. You set everything up again and have a separate server that only you can access that pulls the backups. [rsnapshot](https://rsnapshot.org)

You find out about updates and start applying them weekly. You become bored with doing this weekly and automate applying patches via your package manager. Except for zero-day exploits you're covered. [Unattended-upgrades](https://www.cyberciti.biz/faq/how-to-set-up-automatic-updates-for-ubuntu-linux-18-04/)

Pingdom notification. The server works intermittently. You log in, there's no more disk space. You clean up logs, server is back online. You set up log rotation and start monitoring disk space. You're covered. [logrotate](https://linux.die.net/man/8/logrotate)

Pingdom notification. The server works intermittently. You log in, there is enough disk space. A friend says to check for free inodes. You remove millions of small files from a stupid script you wrote a while ago and start monitoring for inodes. [Inodes](https://www.youtube.com/watch?v=qXVbNlMG28I)

You're still using cronjobs. You learn of monitoring systems. You ask your friend and he says "try Hobbit, it's easy". You set it up and get monitoring and alerting right out of the box. WHOA! Disk, inodes, CPU, memory, files, ports... [Xymon (ex Hobbit)](https://xymon.sourceforge.io)

This new tool allows you to monitor everything and get nice graphs for stuff. You can even monitor stuff of your own, writing small shell scripts. The world is your oyster and you feel awesome. You're a sysadmin. It's still 2010. History is still being written.

If you liked what you read, consider following me for more and retweeting this. I'll be writing about old school system administration, DevOps, Docker, Kubernetes and serverless, infrastructure as code, monitoring and visualizations, AWS and command line productivity and tools.
