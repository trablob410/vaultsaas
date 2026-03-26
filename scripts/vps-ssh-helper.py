"""VPS SSH helper — runs commands on 103.142.137.232 via paramiko."""
import paramiko
import sys
import time

HOST = "103.142.137.232"
USER = "root"
PASS = "Ec0Cloud@123"

def run(cmd, timeout=300):
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(HOST, username=USER, password=PASS, timeout=15)
    stdin, stdout, stderr = ssh.exec_command(cmd, timeout=timeout)
    out = stdout.read().decode()
    err = stderr.read().decode()
    exit_code = stdout.channel.recv_exit_status()
    ssh.close()
    return out, err, exit_code

def scp_upload(local_path, remote_path):
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(HOST, username=USER, password=PASS, timeout=15)
    sftp = ssh.open_sftp()
    sftp.put(local_path, remote_path)
    sftp.close()
    ssh.close()

if __name__ == "__main__":
    cmd = " ".join(sys.argv[1:]) if len(sys.argv) > 1 else "echo hello"
    out, err, code = run(cmd)
    if out:
        print(out)
    if err:
        print(err, file=sys.stderr)
    sys.exit(code)
