#! /bin/bash

# download l_ipp_2019.1.144.tgz from https://software.seek.intel.com/performance-libraries

tar zxfv l_ipp_2019.1.144.tgz

pushd l_ipp_2019.1.144/rpm

sudo rpm --nodeps -Uvh \
	intel-ipp-common-2019.1-144-2019.1-144.noarch.rpm \
	intel-ipp-st-2019.1-144-2019.1-144.x86_64.rpm \
	intel-ipp-st-devel-2019.1-144-2019.1-144.x86_64.rpm

popd

ipp_root="/opt/intel/compilers_and_libraries_2019.1.144/linux/ipp"

sudo install -D -m 644 ippi.pc "${ipp_root}/lib/intel64_lin/pkgconfig/ippi.pc"

echo "${ipp_root}/lib/intel64_lin" | sudo tee /etc/ld.so.conf.d/ipp.conf

sudo /sbin/ldconfig

pip3 install --user git+https://github.com/anight/pkg-config-force-static.git

echo "export PKG_CONFIG_PATH=\"${ipp_root}/lib/intel64_lin/pkgconfig:\$PKG_CONFIG_PATH\"" >> ~/.bash_profile

# use this if you want to link with ipp libs statically
echo "export PKG_CONFIG_FORCE_STATIC=\"ippi\"" >> ~/.bash_profile

