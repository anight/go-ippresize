#! /bin/sh

rpm --nodeps -Uvh \
	intel-ipp-common-2018.4-274-2018.4-274.noarch.rpm \
	intel-ipp-st-2018.4-274-2018.4-274.x86_64.rpm \
	intel-ipp-st-devel-2018.4-274-2018.4-274.x86_64.rpm

ipp_root="/opt/intel/compilers_and_libraries_2018.5.274/linux/ipp"

echo PKG_CONFIG_PATH="${ipp_root}/lib/intel64_lin/pkgconfig:\$PKG_CONFIG_PATH" >> /etc/profile

install -d -m 644 ippi.pc ${ipp_root}/lib/intel64_lin/pkgconfig/

echo "${ipp_root}/lib/intel64_lin" > /etc/ld.so.conf.d/ipp.conf

/sbin/ldconfig
