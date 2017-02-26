#!/bin/bash
cd ~
dd if=/dev/zero of=disk.raw bs=1024k count=2048
echo ">>>>>>>>>>>>>>>>>>>> Created a blank image."

sudo losetup -f disk.raw
sudo losetup -a
sudo mkfs.ext4 /dev/loop0
echo ">>>>>>>>>>>>>>>>>>>> Format the block in ext4."

mkdir mnt
sudo mount /dev/loop0 ~/mnt
echo ">>>>>>>>>>>>>>>>>>>> Mounted to mnt."

sudo git clone https://github.com/istio/istio-testing
sudo mv istio-testing/* ~/mnt
sudo rm -rf istio-testing/
echo ">>>>>>>>>>>>>>>>>>>> Loaded source code to the image."

sudo umount /dev/loop0
sudo losetup -d /dev/loop0
rm -rf mnt

tar -czvf istio-image.tar.gz disk.raw
echo ">>>>>>>>>>>>>>>>>>>> Packaged the image."

export K8S_SCOPES='https://www.googleapis.com/auth/appengine.admin,https://www.googleapis.com/auth/cloud-platform,https://www.googleapis.com/auth/compute,https://www.googleapis.com/auth/devstorage.full_control,https://www.googleapis.com/auth/logging.write,https://www.googleapis.com/auth/projecthosting,https://www.googleapis.com/auth/servicecontrol,https://www.googleapis.com/auth/service.management'

gsutil cp istio-image.tar.gz gs://istio-image/
echo ">>>>>>>>>>>>>>>>>>>> Uploaded to Cloud Storage."
gcloud compute images create jenkins-home-image --source-uri gs://istio-image/istio-image.tar.gz
echo ">>>>>>>>>>>>>>>>>>>> Created a CE image."
gcloud compute disks create jenkins-home --zone us-central1-f --image jenkins-home-image
echo ">>>>>>>>>>>>>>>>>>>> Created a CE disk jenkins-home."
