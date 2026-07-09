%global debug_package %{nil}
%global sname pgedge-docloader

Name:           %{sname}
Version:        %{docloader_version}
Release:        %{docloader_buildnum}%{?dist}
Summary:        pgEdge docloader tool for converting HTML and RST docs into Markdown, and loading them into PostgreSQL

License:        PostgreSQL License
URL:            https://github.com/pgEdge/%{sname}

Source0:	https://github.com/pgEdge/%{sname}/releases/download/v%{docloader_version}/%{sname}_%{docloader_version}_Linux_%{arch}.tar.gz
Source1:        config.yml
Source2:	LICENCE.md

%description
A tool for converting HTML and RST docs into Markdown, and loading them into PostgreSQL

%prep
%setup -q -c -n pgedge-docloader-%{version}
cp %{SOURCE2} .

%build
syft dir:%{_builddir} -o cyclonedx-json > %{_builddir}/%{sname}-sbom.json || exit 1

KEY_ID=$(gpg --list-secret-keys --with-colons | awk -F: '/^sec/{print $5}' | head -n 1); export KEY_ID
gpg --armor --detach-sign --output %{_builddir}/%{sname}-sbom.json.asc %{_builddir}/%{sname}-sbom.json || exit 1

%install
install -D -m 0755 %{sname} %{buildroot}/usr/bin/%{sname}
install -D -m 0644 %{SOURCE1} %{buildroot}%{_sysconfdir}/pgedge/docloader/config.yml
mkdir -p %{buildroot}%{_datadir}/%{sname}
install -p -m 0644 %{_builddir}/%{sname}-sbom.json %{buildroot}%{_datadir}/%{sname}/%{sname}-sbom.json
install -p -m 0644 %{_builddir}/%{sname}-sbom.json.asc %{buildroot}%{_datadir}/%{sname}/%{sname}-sbom.json.asc

%files
%license LICENCE.md
%doc README.md
%{_bindir}/%{sname}
%config(noreplace) %{_sysconfdir}/pgedge/docloader/config.yml
%{_datadir}/%{sname}/%{sname}-sbom.json
%{_datadir}/%{sname}/%{sname}-sbom.json.asc

%changelog
* Fri Mar 13 2026 Muhammad Aqeel <muhammad.aqeel@pgedge.com> - 1.0.0
- Update RPM package of pgedge-docloader
* Mon Dec 15 2025 Muhammad Aqeel <muhammad.aqeel@pgedge.com> - 1.0.0-beta1
- Initial RPM package of pgedge-docloader

