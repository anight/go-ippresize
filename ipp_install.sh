#! /bin/bash

sudo rpm --nodeps -Uvh \
	intel-ipp-common-2018.4-274-2018.4-274.noarch.rpm \
	intel-ipp-st-2018.4-274-2018.4-274.x86_64.rpm \
	intel-ipp-st-devel-2018.4-274-2018.4-274.x86_64.rpm

ipp_root="/opt/intel/compilers_and_libraries_2018.5.274/linux/ipp"

sudo install -D -m 644 ippi.pc "${ipp_root}/lib/intel64_lin/pkgconfig/"

echo "${ipp_root}/lib/intel64_lin" | sudo tee /etc/ld.so.conf.d/ipp.conf

sudo /sbin/ldconfig

pip3 install --user git+https://github.com/anight/pkg-config-force-static.git

echo "export PKG_CONFIG_PATH=\"${ipp_root}/lib/intel64_lin/pkgconfig:\$PKG_CONFIG_PATH\"" >> ~/.bash_profile

# use this if you want to link with ipp libs statically
echo "export PKG_CONFIG_FORCE_STATIC=\"ippi\"" >> ~/.bash_profile

