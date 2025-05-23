# Go Raspberry PI Temperature Monitor

**Go-Raspi-Temp-Monitor** is a temperature monitoring application designed for Raspberry PI devices. It will read the CPU temperature at regular intervals and send alerts via email when the temperature exceeds a specified threshold.

<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/e2815396-a682-4f94-97f7-52abd5345c0b"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/e2815396-a682-4f94-97f7-52abd5345c0b"><img src="[https://github.com/user-attachments/assets/e2815396-a682-4f94-97f7-52abd5345c0b](https://github.com/user-attachments/assets/e2815396-a682-4f94-97f7-52abd5345c0b)" width=500></picture>
</p>

![GitHub Release](https://img.shields.io/github/v/release/richbl/go-raspi-temp-monitor?include_prereleases&sort=semver&display_name=tag&style=flat&color=blue)  [![Go Report Card](https://goreportcard.com/badge/github.com/richbl/go-raspi-temp-monitor)](https://goreportcard.com/report/github.com/richbl/go-raspi-temp-monitor)  [![Codacy Badge](https://app.codacy.com/project/badge/Grade/9578d68b618d4b2a8e1dd928b8ebb9d6)](https://app.codacy.com/gh/richbl/go-raspi-temp-monitor/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) 
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=richbl_go-raspi-temp-monitor&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=richbl_go-raspi-temp-monitor)

## Rationale

I run a number of Raspberry PI devices as part of my [Distributed Motion Surveillance Security System (DMS3) project](https://github.com/richbl/go-distributed-motion-s3) and noticed that every so often, one of them would just quietly drop from the network. Reviewing the logs of the dead devices wasn't providing anything definitive, so I thought that--perhaps--the issue may relate to unusually high CPU temperatures. In short, I wanted a simple way to remotely monitor these device temperatures.

Go-Raspi-Temp-Monitor does just this: it's a very simple project. Once installed on the target Raspberry PI device, it will read the CPU temperature at regular intervals and send alerts via email when the temperature exceeds a specified threshold.

## Requirements

The **Go-Raspi-Temp-Monitor** project sources will first need to be compiled on a build machine, and then moved to a target Raspberry PI device.

### Build Machine Requirements

As this project was written in Go (1.23.5), the build machine requires a recent version of Go installed and configured.

### Target Device Requirements

The target device requires the following:

- A recent version of the [Raspberry PI OS](https://www.raspberrypi.com/software/operating-systems/). This project was written for Raspberry PI devices running Debian Bookworm (32-bit), but should work on other Raspberry PI OS variants as well
- The [`mailutils`](https://www.gnu.org/software/mailutils/) package installed and configured on the target device
    - It's highly recommended to install the [`ssmtp`](https://wiki.debian.org/sSMTP) package as well, which provides support for SMTP authentication (useful for Gmail account authentication)

## Installation

The process of installing **Go-Raspi-Temp-Monitor** is as follows:

1. Clone the **Go-Raspi-Temp-Monitor** repository
2. Build the application on the build machine
3. Install the OS-specific binary application on the target Raspberry PI device
4. Optionally (recommended), configure the application to run as a systemd service

### 1. Cloning the **Go-Raspi-Temp-Monitor** Repository

To clone the **Go-Raspi-Temp-Monitor** repository, run the following command on the build machine:

```console
git clone https://github.com/richbl/go-raspi-temp-monitor.git
```

Once cloned, the **Go-Raspi-Temp-Monitor** application can be built and installed on the target Raspberry PI device.

> Note: the `main.go` file in the project includes a number of constants that can be edited to configure the application. It would be worthwhile to review these constants before building the application (i.e., the location of the `mail` command, which defaults to `/usr/bin/mail`).

### 2. Building **Go-Raspi-Temp-Monitor**

Building the **Go-Raspi-Temp-Monitor** application requires an understanding of the target Raspberry PI architecture and operating system, as the build process will be a cross-compile, meaning that the resulting binary will be compatible with the target architecture (and not necessarily the build machine architecture).

The topic of cross-compiling is beyond the scope of this document, but the [Go documentation](https://pkg.go.dev/cmd/go#hdr-Environment_variables) provides a good overview of the process. As well, a good overview of the relevant cross-compiling flags used in this project--`GOOS`, `GOARCH`, and `GOARM`--can be found [here](https://go.dev/doc/install/source#environment).

For this project, the target devices are Raspberry PI computers running the 32-bit version of the [Raspberry PI OS](https://www.raspberrypi.com/software/operating-systems/).

To build the application, run the following command from the project root directory on the build machine:

```console
GOOS=linux GOARCH=arm GOARM=7 go build -o go-raspi-temp-monitor cmd/main.go
```

> Note that `GOARM=6` may be required when building the binary for older Raspberry PI devices (e.g., Raspberry PI Zero). Check your device specifications to determine the correct value required

This will result in a binary file named `go-raspi-temp-monitor` that can then be installed on the target Raspberry PI device.

#### Building for Raspberry PI OS 64-bit

To build the application for a 64-bit version of the Raspberry PI OS, run the following command from the project root directory on the build machine:

```console
GOOS=linux GOARCH=arm64 go build -o go-raspi-temp-monitor cmd/main.go
```

### 3. Installing **Go-Raspi-Temp-Monitor**

Once built, the **Go-Raspi-Temp-Monitor** application can be transferred over to the target Raspberry PI device.

By convention, the application binary, `go-raspi-temp-monitor`, should be copied into the `/usr/local/bin` directory.

### 4. Configuring **Go-Raspi-Temp-Monitor** as a Systemd Service

Once installed on the target Raspberry PI device, **Go-Raspi-Temp-Monitor** can be configured to run as a systemd service.  

To configure **Go-Raspi-Temp-Monitor** to run as a systemd service:

1. Edit the **Go-Raspi-Temp-Monitor** service file (`go-raspi-temp-monitor.service`), making sure to replace the placeholder email address (`your_email@example.com`) with an appropriate email address. As well, be sure to review and edit the command line flags passed into the application (the default values will work for most users)
2. Copy the **Go-Raspi-Temp-Monitor** service file (`go-raspi-temp-monitor.service`) into the `/etc/systemd/system` directory
3. Enable the **Go-Raspi-Temp-Monitor** service. This can be done via the command:

```console
sudo systemctl enable go-raspi-temp-monitor.service
```

4. Start the **Go-Raspi-Temp-Monitor** service. This can be done via the command

```console
sudo systemctl start go-raspi-temp-monitor.service
```

At this point, **Go-Raspi-Temp-Monitor** should be running as a systemd service on the target Raspberry PI device. To check the status of the service, use the following command:

```console
sudo systemctl status go-raspi-temp-monitor.service
```

### Testing Email Operation and Delivery

Optionally, and to confirm that the email delivery feature is working as expected, **Go-Raspi-Temp-Monitor** can be tested by sending a real-time email with the following command:

```console
go-raspi-temp-monitor -test-email -recipient=your_email@example.com
```

The application will respond with:

```console
2025/05/15 17:39:22 ----- Starting Go-Raspi-Temp-Monitor 0.7.0
2025/05/15 17:39:22 ----- Configuration
2025/05/15 17:39:22 |
2025/05/15 17:39:22 | Temperature threshold ('-threshold'): 60.00°C
2025/05/15 17:39:22 | Check interval ('-interval'): 5m0s
2025/05/15 17:39:22 | Email recipient ('-recipient'): your_email@example.com
2025/05/15 17:39:22 | Mail command: /usr/bin/mail
2025/05/15 17:39:22 | Device hostname: picam-alpha
2025/05/15 17:39:22 |
2025/05/15 17:39:22 Attempting to send email to your_email@example.com
2025/05/15 17:39:27 Email sent successfully to your_email@example.com
2025/05/15 17:39:27 ----- Exiting Go-Raspi-Temp-Monitor 0.7.0
```
A test email should arrive entitled "Go-Raspi-Temp-Monitor: Test Alert (`<hostname>`)" with the following content:

```console
Warning: this is a test email
Hostname: <hostname>
Current CPU temperature: 47.24°C
Timestamp: Thu, 15 May 2025 18:02:57 PDT
```

## Usage

### Command Line Flags

For portability, the **Go-Raspi-Temp-Monitor** application does not use a configuration file. Instead, command line flags are used to directly configure the application. These flags are as follows:

- `-recipient`: Recipient email address for alert notifications
- `-threshold`: CPU temperature (Celsius) threshold to reach before sending an alert
- `-interval`: Interval for checking CPU temperature (e.g., '5m', '1h')
- `-test-email`: Immediately send a test email and then exit (requires -recipient to be set)
- `-help`: Displays this help message and then exits

While it's likely more typical that users will use the **Go-Raspi-Temp-Monitor** application as a systemd service, the application can be run at any time from the command line.

As an example, the following command will set the recipient email address to `your_email@example.com`, set the CPU temperature threshold to 60 degrees Celsius, and set the check interval to 5 seconds:

```console
go-raspi-temp-monitor -recipient=your_email@example.com -threshold=60 -interval=5s
```

The output of this command is as follows:

```console
2025/05/15 17:47:07 ----- Starting Go-Raspi-Temp-Monitor 0.7.0
2025/05/15 17:47:07 ----- Configuration
2025/05/15 17:47:07 |
2025/05/15 17:47:07 | Temperature threshold ('-threshold'): 60.00°C
2025/05/15 17:47:07 | Check interval ('-interval'): 5s
2025/05/15 17:47:07 | Email recipient ('-recipient'): your_email@example.com
2025/05/15 17:47:07 | Mail command: /usr/bin/mail
2025/05/15 17:47:07 | Device hostname: picam-alpha
2025/05/15 17:47:07 |
2025/05/15 17:47:07 ----- Monitoring
2025/05/15 17:47:07 Current CPU temperature: 42.00°C
2025/05/15 17:47:12 Current CPU temperature: 42.00°C
2025/05/15 17:47:17 Current CPU temperature: 42.00°C
2025/05/15 17:47:19 Received signal interrupt: shutting down
2025/05/15 17:47:19 ----- Exiting Go-Raspi-Temp-Monitor 0.7.0
```

In the above example, if the CPU temperature ever exceeds 60 degrees Celsius, the application would send an email notification to `your_email@example.com`.

> Note that, to exit the application, you use the `Ctrl+C` (break) key combination, and the application will shut down and exit gracefully.

## Roadmap

At the moment, there's not much of a roadmap to consider. In general, this is a executable doing some pretty basic stuff.

That said, it might be useful to incorporate a full email package into this codebase rather than relying on an external mailer solution (e.g., `mailutils` and `ssmtp` packages).

Regardless, if you have any thoughts or ideas for improvement, send them my way.

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/richbl/go-raspi-temp-monitor/blob/main/LICENSE) file for details.
