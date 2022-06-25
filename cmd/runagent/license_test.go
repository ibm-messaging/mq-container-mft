/*
© Copyright IBM Corporation 2020

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE_2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"os"
	"testing"
)

var licenseTests = []struct {
	in  string
	out string
}{
	{"en_US.UTF_8", "Lic_en.txt"},
	{"en_US.ISO-8859-15", "Lic_en.txt"},
	{"es_GB", "Lic_es.txt"},
	{"el_ES.UTF_8", "Lic_el.txt"},
	// Cover a wide variety of valid values
	{"af", "Lic_en.txt"},
	{"af_ZA", "Lic_en.txt"},
	{"ar", "Lic_en.txt"},
	{"ar_AE", "Lic_en.txt"},
	{"ar_BH", "Lic_en.txt"},
	{"ar_DZ", "Lic_en.txt"},
	{"ar_EG", "Lic_en.txt"},
	{"ar_IQ", "Lic_en.txt"},
	{"ar_JO", "Lic_en.txt"},
	{"ar_KW", "Lic_en.txt"},
	{"ar_LB", "Lic_en.txt"},
	{"ar_LY", "Lic_en.txt"},
	{"ar_MA", "Lic_en.txt"},
	{"ar_OM", "Lic_en.txt"},
	{"ar_QA", "Lic_en.txt"},
	{"ar_SA", "Lic_en.txt"},
	{"ar_SY", "Lic_en.txt"},
	{"ar_TN", "Lic_en.txt"},
	{"ar_YE", "Lic_en.txt"},
	{"az", "Lic_en.txt"},
	{"az_AZ", "Lic_en.txt"},
	{"az_AZ", "Lic_en.txt"},
	{"be", "Lic_en.txt"},
	{"be_BY", "Lic_en.txt"},
	{"bg", "Lic_en.txt"},
	{"bg_BG", "Lic_en.txt"},
	{"bs_BA", "Lic_en.txt"},
	{"ca", "Lic_en.txt"},
	{"ca_ES", "Lic_en.txt"},
	{"cs", "Lic_cs.txt"},
	{"cs_CZ", "Lic_cs.txt"},
	{"csb_PL", "Lic_en.txt"},
	{"cy", "Lic_en.txt"},
	{"cy_GB", "Lic_en.txt"},
	{"da", "Lic_en.txt"},
	{"da_DK", "Lic_en.txt"},
	{"de", "Lic_de.txt"},
	{"de_AT", "Lic_de.txt"},
	{"de_CH", "Lic_de.txt"},
	{"de_DE", "Lic_de.txt"},
	{"de_LI", "Lic_de.txt"},
	{"de_LU", "Lic_de.txt"},
	{"dv", "Lic_en.txt"},
	{"dv_MV", "Lic_en.txt"},
	{"el", "Lic_el.txt"},
	{"el_GR", "Lic_el.txt"},
	{"en", "Lic_en.txt"},
	{"en_AU", "Lic_en.txt"},
	{"en_BZ", "Lic_en.txt"},
	{"en_CA", "Lic_en.txt"},
	{"en_CB", "Lic_en.txt"},
	{"en_GB", "Lic_en.txt"},
	{"en_IE", "Lic_en.txt"},
	{"en_JM", "Lic_en.txt"},
	{"en_NZ", "Lic_en.txt"},
	{"en_PH", "Lic_en.txt"},
	{"en_TT", "Lic_en.txt"},
	{"en_US", "Lic_en.txt"},
	{"en_ZA", "Lic_en.txt"},
	{"en_ZW", "Lic_en.txt"},
	{"eo", "Lic_en.txt"},
	{"es", "Lic_es.txt"},
	{"es_AR", "Lic_es.txt"},
	{"es_BO", "Lic_es.txt"},
	{"es_CL", "Lic_es.txt"},
	{"es_CO", "Lic_es.txt"},
	{"es_CR", "Lic_es.txt"},
	{"es_DO", "Lic_es.txt"},
	{"es_EC", "Lic_es.txt"},
	{"es_ES", "Lic_es.txt"},
	{"es_ES", "Lic_es.txt"},
	{"es_GT", "Lic_es.txt"},
	{"es_HN", "Lic_es.txt"},
	{"es_MX", "Lic_es.txt"},
	{"es_NI", "Lic_es.txt"},
	{"es_PA", "Lic_es.txt"},
	{"es_PE", "Lic_es.txt"},
	{"es_PR", "Lic_es.txt"},
	{"es_PY", "Lic_es.txt"},
	{"es_SV", "Lic_es.txt"},
	{"es_UY", "Lic_es.txt"},
	{"es_VE", "Lic_es.txt"},
	{"et", "Lic_en.txt"},
	{"et_EE", "Lic_en.txt"},
	{"eu", "Lic_en.txt"},
	{"eu_ES", "Lic_en.txt"},
	{"fa", "Lic_en.txt"},
	{"fa_IR", "Lic_en.txt"},
	{"fi", "Lic_en.txt"},
	{"fi_FI", "Lic_en.txt"},
	{"fo", "Lic_en.txt"},
	{"fo_FO", "Lic_en.txt"},
	{"fr", "Lic_fr.txt"},
	{"fr_BE", "Lic_fr.txt"},
	{"fr_CA", "Lic_fr.txt"},
	{"fr_CH", "Lic_fr.txt"},
	{"fr_FR", "Lic_fr.txt"},
	{"fr_LU", "Lic_fr.txt"},
	{"fr_MC", "Lic_fr.txt"},
	{"gl", "Lic_en.txt"},
	{"gl_ES", "Lic_en.txt"},
	{"gu", "Lic_en.txt"},
	{"gu_IN", "Lic_en.txt"},
	{"he", "Lic_en.txt"},
	{"he_IL", "Lic_en.txt"},
	{"hi", "Lic_en.txt"},
	{"hi_IN", "Lic_en.txt"},
	{"hr", "Lic_en.txt"},
	{"hr_BA", "Lic_en.txt"},
	{"hr_HR", "Lic_en.txt"},
	{"hu", "Lic_en.txt"},
	{"hu_HU", "Lic_en.txt"},
	{"hy", "Lic_en.txt"},
	{"hy_AM", "Lic_en.txt"},
	{"id", "Lic_id.txt"},
	{"id_ID", "Lic_id.txt"},
	{"is", "Lic_en.txt"},
	{"is_IS", "Lic_en.txt"},
	{"it", "Lic_it.txt"},
	{"it_CH", "Lic_it.txt"},
	{"it_IT", "Lic_it.txt"},
	{"ja", "Lic_ja.txt"},
	{"ja_JP", "Lic_ja.txt"},
	{"ka", "Lic_en.txt"},
	{"ka_GE", "Lic_en.txt"},
	{"kk", "Lic_en.txt"},
	{"kk_KZ", "Lic_en.txt"},
	{"kn", "Lic_en.txt"},
	{"kn_IN", "Lic_en.txt"},
	{"ko", "Lic_ko.txt"},
	{"ko_KR", "Lic_ko.txt"},
	{"kok", "Lic_en.txt"},
	{"kok_IN", "Lic_en.txt"},
	{"ky", "Lic_en.txt"},
	{"ky_KG", "Lic_en.txt"},
	{"lt", "Lic_lt.txt"},
	{"lt_LT", "Lic_lt.txt"},
	{"lv", "Lic_en.txt"},
	{"lv_LV", "Lic_en.txt"},
	{"mi", "Lic_en.txt"},
	{"mi_NZ", "Lic_en.txt"},
	{"mk", "Lic_en.txt"},
	{"mk_MK", "Lic_en.txt"},
	{"mn", "Lic_en.txt"},
	{"mn_MN", "Lic_en.txt"},
	{"mr", "Lic_en.txt"},
	{"mr_IN", "Lic_en.txt"},
	{"ms", "Lic_en.txt"},
	{"ms_BN", "Lic_en.txt"},
	{"ms_MY", "Lic_en.txt"},
	{"mt", "Lic_en.txt"},
	{"mt_MT", "Lic_en.txt"},
	{"nb", "Lic_en.txt"},
	{"nb_NO", "Lic_en.txt"},
	{"nl", "Lic_en.txt"},
	{"nl_BE", "Lic_en.txt"},
	{"nl_NL", "Lic_en.txt"},
	{"nn_NO", "Lic_en.txt"},
	{"ns", "Lic_en.txt"},
	{"ns_ZA", "Lic_en.txt"},
	{"pa", "Lic_en.txt"},
	{"pa_IN", "Lic_en.txt"},
	{"pl", "Lic_pl.txt"},
	{"pl_PL", "Lic_pl.txt"},
	{"ps", "Lic_en.txt"},
	{"ps_AR", "Lic_en.txt"},
	{"pt", "Lic_pt.txt"},
	{"pt_BR", "Lic_pt.txt"},
	{"pt_PT", "Lic_pt.txt"},
	{"qu", "Lic_en.txt"},
	{"qu_BO", "Lic_en.txt"},
	{"qu_EC", "Lic_en.txt"},
	{"qu_PE", "Lic_en.txt"},
	{"ro", "Lic_en.txt"},
	{"ro_RO", "Lic_en.txt"},
	{"ru", "Lic_ru.txt"},
	{"ru_RU", "Lic_ru.txt"},
	{"sa", "Lic_en.txt"},
	{"sa_IN", "Lic_en.txt"},
	{"se", "Lic_en.txt"},
	{"se_FI", "Lic_en.txt"},
	{"se_FI", "Lic_en.txt"},
	{"se_FI", "Lic_en.txt"},
	{"se_NO", "Lic_en.txt"},
	{"se_NO", "Lic_en.txt"},
	{"se_NO", "Lic_en.txt"},
	{"se_SE", "Lic_en.txt"},
	{"se_SE", "Lic_en.txt"},
	{"se_SE", "Lic_en.txt"},
	{"sk", "Lic_en.txt"},
	{"sk_SK", "Lic_en.txt"},
	{"sl", "Lic_sl.txt"},
	{"sl_SI", "Lic_sl.txt"},
	{"sq", "Lic_en.txt"},
	{"sq_AL", "Lic_en.txt"},
	{"sr_BA", "Lic_en.txt"},
	{"sr_BA", "Lic_en.txt"},
	{"sr_SP", "Lic_en.txt"},
	{"sr_SP", "Lic_en.txt"},
	{"sv", "Lic_en.txt"},
	{"sv_FI", "Lic_en.txt"},
	{"sv_SE", "Lic_en.txt"},
	{"sw", "Lic_en.txt"},
	{"sw_KE", "Lic_en.txt"},
	{"syr", "Lic_en.txt"},
	{"syr_SY", "Lic_en.txt"},
	{"ta", "Lic_en.txt"},
	{"ta_IN", "Lic_en.txt"},
	{"te", "Lic_en.txt"},
	{"te_IN", "Lic_en.txt"},
	{"th", "Lic_en.txt"},
	{"th_TH", "Lic_en.txt"},
	{"tl", "Lic_en.txt"},
	{"tl_PH", "Lic_en.txt"},
	{"tn", "Lic_en.txt"},
	{"tn_ZA", "Lic_en.txt"},
	{"tr", "Lic_tr.txt"},
	{"tr_TR", "Lic_tr.txt"},
	{"tt", "Lic_en.txt"},
	{"tt_RU", "Lic_en.txt"},
	{"ts", "Lic_en.txt"},
	{"uk", "Lic_en.txt"},
	{"uk_UA", "Lic_en.txt"},
	{"ur", "Lic_en.txt"},
	{"ur_PK", "Lic_en.txt"},
	{"uz", "Lic_en.txt"},
	{"uz_UZ", "Lic_en.txt"},
	{"uz_UZ", "Lic_en.txt"},
	{"vi", "Lic_en.txt"},
	{"vi_VN", "Lic_en.txt"},
	{"xh", "Lic_en.txt"},
	{"xh_ZA", "Lic_en.txt"},
	{"zh", "Lic_zh.txt"},
	{"zh_CN", "Lic_zh.txt"},
	{"zh_HK", "Lic_zh.txt"},
	{"zh_MO", "Lic_zh.txt"},
	{"zh_SG", "Lic_zh.txt"},
	{"zh_TW", "Lic_zh_tw.txt"},
	{"zu", "Lic_en.txt"},
	{"zu_ZA", "Lic_en.txt"},
}

func TestResolveLicenseFile(t *testing.T) {
	for _, table := range licenseTests {
		os.Setenv("LANG", table.in)
		f := resolveLicenseFile()
		if f != table.out {
			t.Errorf("resolveLicenseFile() with LANG=%v - expected %v, got %v", table.in, table.out, f)
		}
	}
}
