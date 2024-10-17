# Sicher - A Simple Linux Backup Daemon

## Overview

**Sicher** is a Linux-based backup daemon designed to automatically back up files from a specified source directory to a destination directory. It allows for customization of the backup process via a configuration file, ensuring that large files, restricted directories, and specific file extensions can be skipped. The tool runs as a background process (daemon) and supports logging for tracking backup operations.

## Features

- Backup source directory (`srcDir`) to destination directory (`dstDir`).
- Configuration via a TOML file (`sicher.toml`).
- Adjustable worker concurrency for copying files.
- Skips restricted directories, files, and file extensions.
- Logs backup events and errors in `sicher.log`.
- Uses a rolling log with compression and limited size management.

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Logging](#logging)
- [Stopping Sicher](#stopping-sicher)

## Installation

1. **Clone the repository**:

    ```bash
    git clone https://github.com/scott-mescudi/sicher.git
    cd sicher
    ```

2. **Install dependencies**:

   Ensure Go is installed on your system. Install dependencies by running:

    ```bash
    go mod tidy
    ```

3. **Build the project**:

    ```bash
    go build -o sicher
    ```

4. **Create a configuration file**:

    Sicher reads its settings from a configuration file named `sicher.toml`. Here's an example file structure:

    ```toml
    srcDir = "/path/to/source"
    dstDir = "/path/to/destination"
    maxWorkers = 4
    memUsage = 104857600 
    backupFreq = 60 
    maxFileSize = 2147483648 
    
    [restrictedDirs]
    temp = true
    logs = true

    [restrictedFiles]
    ".DS_Store" = true
    "Thumbs.db" = true

    [restrictedExtensions]
    ".tmp" = true
    ".log" = true
    ```

   Customize the paths and restrictions according to your needs.

5. **Run Sicher**:

    To run Sicher, execute:

    ```bash
    ./sicher
    ```

   Sicher will start as a background process and begin backing up files based on the configuration.

## Configuration

The configuration file (`sicher.toml`) allows you to define various settings for Sicher:

- **`srcDir`**: Path to the source directory to back up.
- **`dstDir`**: Path to the destination directory where the backup will be stored.
- **`maxWorkers`**: Number of concurrent workers to process file copies.
- **`memUsage`**: Memory buffer size (in bytes) for file copying operations.
- **`backupFreq`**: Backup frequency in minutes.
- **`maxFileSize`**: Maximum file size (in bytes) to back up.
- **`restrictedDirs`**: List of directories to skip during the backup.
- **`restrictedFiles`**: List of specific files to skip.
- **`restrictedExtensions`**: List of file extensions to skip.

### Example Configuration

```toml
  srcDir = "/path/to/source"
  dstDir = "/path/to/destination"
  maxWorkers = 4
  memUsage = 104857600 
  backupFreq = 60 
  maxFileSize = 2147483648 
  
  [restrictedDirs]
  temp = true
  logs = true

  [restrictedFiles]
  ".DS_Store" = true
  "Thumbs.db" = true

  [restrictedExtensions]
  ".tmp" = true
  ".log" = true
```

## Usage

Once installed and configured, Sicher can be run simply by executing the compiled binary:

```bash
./sicher
```

The daemon will read the configuration, start the backup process immediately, and continue at regular intervals defined by `backupFreq`.

### Automatic Backup

Sicher will check the source directory and copy files that meet the following criteria:
- Not restricted by the configuration.
- Not larger than the `maxFileSize`.
- Modified since the last backup.

### Handling Large Directories

The number of workers (threads) that Sicher uses to copy files concurrently can be adjusted using the `maxWorkers` option in the configuration file. This is particularly useful for large directories to speed up the backup process.

## Logging

Sicher generates logs to track the progress of backup operations. The logs are managed by the [lumberjack](https://github.com/natefinch/lumberjack) package for automatic rotation and compression. 

- **Log Location**: By default, the log file `sicher.log` is created in the same directory where the `sicher` binary is executed. To ensure consistent log management, you may want to configure a dedicated log directory, such as `/var/log/sicher.log`.
  
- **Log Management**: The log file rotates when it reaches 10 MB, keeping up to 3 backup logs compressed. This helps manage disk space while retaining sufficient log history for review.

- **Log Content**: The log records important events, including successful backups, errors encountered during the backup process, and any skipped files or directories due to restrictions.

- **Example Log Output**:
    ```
    [2024-10-17 14:22:01] INFO: Backup started for /path/to/source
    [2024-10-17 14:22:05] INFO: Backed up file: /path/to/source/example.txt
    [2024-10-17 14:22:10] ERROR: Failed to copy file: /path/to/source/temp/example.tmp
    ```

- **Monitoring Logs**: Since we specified `SyslogIdentifier=sicher-backup`, Sicher’s output will also be available in the system logs. You can monitor the Sicher logs by using the `journalctl` command:
    ```bash
    sudo journalctl -u sicher.service -f
    ```


## Stopping Sicher

To stop Sicher, you can send a `SIGINT` or `SIGTERM` signal, for example, by pressing `Ctrl+C` or using the `kill` command.

```bash
pkill -SIGINT sicher
```

This will gracefully stop the daemon and clean up running workers.

## Running Sicher as a Systemd Service

To ensure that **Sicher** runs automatically on system boot as a daemon, you can create a systemd service for it. Follow these steps to set up and manage the Sicher service on your Linux system:

### 1. Move the Sicher Binary to a System Directory

First, move the compiled `sicher` binary to a location accessible by all users, such as `/usr/local/bin`.

```bash
sudo mv sicher /usr/local/bin/
```

### 2. Create a Systemd Service File

Next, create a new systemd service file for Sicher in the `/etc/systemd/system` directory:

```bash
sudo nano /etc/systemd/system/sicher.service
```

Add the following content to the file:

```ini
[Unit]
Description=Sicher Backup Daemon
After=network.target

[Service]
ExecStart=/usr/local/bin/sicher
Restart=on-failure
User=root
WorkingDirectory=/usr/local/bin
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=sicher-backup
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```

### Explanation of the service file options:

- **`ExecStart`**: Specifies the path to the Sicher binary.
- **`Restart=on-failure`**: Automatically restarts Sicher if it crashes.
- **`User=root`**: Runs Sicher as the `root` user (you can change this to another user if needed).
- **`WorkingDirectory`**: Sets the working directory for Sicher.
- **`SyslogIdentifier=sicher-backup`**: Ensures Sicher logs to the system log under the identifier `sicher-backup`.
- **`LimitNOFILE=4096`**: Sets a file descriptor limit to avoid file limitations during backup operations.
  
### 3. Reload systemd and Enable the Service

After creating the service file, reload the systemd manager configuration to register the new service:

```bash
sudo systemctl daemon-reload
```

Enable the Sicher service to start automatically at boot:

```bash
sudo systemctl enable sicher
```

### 4. Start the Sicher Service

Now, start the Sicher service manually to ensure it works as expected:

```bash
sudo systemctl start sicher
```

Check the status of the service to verify that it's running:

```bash
sudo systemctl status sicher
```

### 5. Monitor Logs

Since we specified `SyslogIdentifier=sicher-backup`, Sicher’s output will be available in the system logs. You can monitor the Sicher logs by using the `journalctl` command:

```bash
sudo journalctl -u sicher.service -f
```


### 6. Stopping and Restarting the Service

You can stop or restart the Sicher service anytime with the following commands:

- To stop Sicher:

  ```bash
  sudo systemctl stop sicher
  ```

- To restart Sicher:

  ```bash
  sudo systemctl restart sicher
  ```

### 7. Confirm Sicher Runs on Boot

Reboot your system to confirm that Sicher starts automatically:

```bash
sudo reboot
```

Once the system is back online, you can check the status of the Sicher daemon:

```bash
sudo systemctl status sicher
```

---

With these steps, **Sicher** is now set up to run automatically on boot, ensuring your backup daemon is always operational!

## Contributing

Feel free to fork the repository, submit issues, or send pull requests to contribute to the project. For more details, refer to the repository on GitHub: [scott-mescudi/sicher](https://github.com/scott-mescudi/sicher).

---

Thank you for using **Sicher**!