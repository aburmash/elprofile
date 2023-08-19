
all: elprofile

elprofile:
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-extldflags -static" -o elprofile cmd/elprofile/main.go

ubi-profiles: ubi90.profile ubi91.profile ubi92.profile


ubi88.profile: elprofile
	apptainer exec docker://redhat/ubi8:8.8 ./elprofile -g > ubi88.profile

ubi91.profile: elprofile
	apptainer exec docker://redhat/ubi9:9.1 ./elprofile -g > ubi91.profile

ubi92.profile: elprofile
	apptainer exec docker://redhat/ubi9:9.2 ./elprofile -g > ubi92.profile


test-rocky88-container: ubi88.profile
	apptainer exec docker://rockylinux:8.8 ./elprofile -q ubi88.profile 

test-rocky91-container: ubi91.profile
	apptainer exec docker://rockylinux:9.1 ./elprofile -q ubi92.profile 

test-rocky92-container: ubi92.profile
	apptainer exec docker://rockylinux:9.2 ./elprofile -q ubi92.profile 


test-alma88-container: ubi88.profile
	apptainer exec docker://almalinux:8.8 ./elprofile -q ubi88.profile 

test-alma91-container: ubi91.profile
	apptainer exec docker://almalinux:9.1 ./elprofile -q ubi92.profile 

test-alma92-container: ubi92.profile
	apptainer exec docker://almalinux:9.2 ./elprofile -q ubi92.profile 


clean:
	rm -f elprofile

clean-all: clean
	rm -f ubi90.profile
	rm -f ubi91.profile
	rm -f ubi92.profile
