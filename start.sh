#!/bin/bash
set -e

# Start local MariaDB/MySQL if installed and not already running
if command -v mysqld_safe >/dev/null 2>&1; then
    if ! mysqladmin ping -h 127.0.0.1 -u seonaut -pseonaut --silent 2>/dev/null; then
        echo "Starting local MariaDB daemon..."
        mkdir -p /run/mysqld /var/lib/mysql
        chown -R mysql:mysql /run/mysqld /var/lib/mysql 2>/dev/null || true
        /usr/bin/mysqld_safe --user=mysql --skip-syslog >/dev/null 2>&1 &
        
        # Wait up to 10 seconds for MariaDB to respond
        for i in {1..20}; do
            if mysqladmin ping -h 127.0.0.1 -u seonaut -pseonaut --silent 2>/dev/null; then
                echo "MariaDB is ready."
                break
            fi
            sleep 0.5
        done
    fi
fi

# Ensure seonaut binary is executable
if [ -f "./seonaut" ]; then
    chmod +x ./seonaut
fi

# Execute seonaut with passed arguments
exec ./seonaut -c config "$@"
