import { useState, useEffect } from 'react';
import { getSettings, updateSettings, testSmtp, syncHexoPosts, getHexoSyncStatus } from '../../api/setting';
import { toast } from 'sonner';
import { Settings as SettingsIcon, RefreshCw, Mail, Save } from 'lucide-react';
import CustomSelect from '../../components/CustomSelect';

const smtpEncryptionOptions = [
  { value: 'tls', label: 'TLS（端口 587 常用）' },
  { value: 'ssl', label: 'SSL（端口 465 常用）' },
  { value: 'none', label: '无加密' },
];

const tabs = [
  { id: 'general', label: '基础设置' },
  { id: 'seo', label: 'SEO 设置' },
  { id: 'smtp', label: '邮件服务 (SMTP)' },
  { id: 'hexo', label: 'Hexo 同步' },
];

export default function Settings() {
  const [activeTab, setActiveTab] = useState('general');
  const [settings, setSettings] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [testEmail, setTestEmail] = useState('');
  const [syncStatus, setSyncStatus] = useState<any>(null);

  const fetchSettings = async (group: string) => {
    setLoading(true);
    try {
      const res = await getSettings(group);
      if (res.code === 200) {
        setSettings(res.data.settings || {});
      } else {
        toast.error(res.message || '获取设置失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求设置失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchSyncStatus = async () => {
    try {
      const res = await getHexoSyncStatus();
      if (res.code === 200) {
        setSyncStatus(res.data);
      }
    } catch {
      // 静默失败
    }
  };

  useEffect(() => {
    if (activeTab === 'hexo') {
      setLoading(true);
      void (async () => {
        try {
          const [setRes, syncRes] = await Promise.all([getSettings('hexo'), getHexoSyncStatus()]);
          if (setRes.code === 200) {
            setSettings(setRes.data.settings || {});
          } else {
            toast.error(setRes.message || '获取 Hexo 设置失败');
          }
          if (syncRes.code === 200) {
            setSyncStatus(syncRes.data);
          }
        } catch (e: any) {
          toast.error(e.response?.data?.message || '加载 Hexo 页失败');
        } finally {
          setLoading(false);
        }
      })();
    } else {
      void fetchSettings(activeTab);
    }
  }, [activeTab]);

  const handleSave = async () => {
    setSaving(true);
    try {
      const payload =
        activeTab === 'hexo'
          ? Object.fromEntries(Object.entries(settings).filter(([k]) => k !== 'hexo.hexo_dir'))
          : settings;
      const res = await updateSettings(activeTab, payload);
      if (res.code === 200) {
        toast.success('保存成功');
        if (activeTab === 'hexo') {
          setSettings(res.data.settings || {});
        }
      } else {
        toast.error(res.message || '保存失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '保存请求失败');
    } finally {
      setSaving(false);
    }
  };

  const handleTestSmtp = async () => {
    if (!testEmail) {
      toast.error('请输入测试接收邮箱');
      return;
    }
    try {
      const res = await testSmtp(testEmail);
      if (res.code === 200) {
        toast.success('测试邮件发送成功');
      } else {
        toast.error(res.message || '测试邮件发送失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '测试请求失败');
    }
  };

  const handleSyncHexo = async () => {
    try {
      const res = await syncHexoPosts(true);
      if (res.code === 200) {
        toast.success('同步成功');
        fetchSyncStatus();
      } else {
        toast.error(res.message || '同步失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '同步请求失败');
    }
  };

  const handleChange = (key: string, value: string) => {
    setSettings(prev => ({ ...prev, [key]: value }));
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <SettingsIcon className="h-5 w-5 shrink-0 text-muted-foreground" aria-hidden />
        <h2 className="text-xl font-semibold tracking-tight text-foreground">系统设置</h2>
      </div>

      <div className="flex flex-col md:flex-row gap-6">
        <div className="w-full md:w-48 flex-shrink-0">
          <nav className="flex flex-col space-y-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`px-3 py-2 text-sm font-medium rounded-md text-left transition-colors ${
                  activeTab === tab.id
                    ? 'bg-primary/10 text-primary'
                    : 'text-foreground hover:bg-accent'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        <div className="flex-1 bg-card border border-border rounded-lg p-6">
          {loading ? (
            <div className="py-12 text-center text-muted-foreground">加载中...</div>
          ) : (
            <div className="space-y-6 max-w-2xl">
              {activeTab === 'general' && (
                <>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">站点名称</label>
                    <input
                      type="text"
                      value={settings['site_name'] || ''}
                      onChange={(e) => handleChange('site_name', e.target.value)}
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">站点 URL</label>
                    <input
                      type="url"
                      value={settings['site_url'] || ''}
                      onChange={(e) => handleChange('site_url', e.target.value)}
                      placeholder="https://example.com"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">站点描述</label>
                    <textarea
                      value={settings['site_description'] || ''}
                      onChange={(e) => handleChange('site_description', e.target.value)}
                      rows={3}
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                </>
              )}

              {activeTab === 'seo' && (
                <>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">SEO 标题</label>
                    <input
                      type="text"
                      value={settings['seo_title'] || ''}
                      onChange={(e) => handleChange('seo_title', e.target.value)}
                      placeholder="默认使用站点名称"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">SEO 描述</label>
                    <textarea
                      value={settings['seo_description'] || ''}
                      onChange={(e) => handleChange('seo_description', e.target.value)}
                      rows={3}
                      placeholder="默认使用站点描述"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">关键词</label>
                    <input
                      type="text"
                      value={settings['site_keywords'] || ''}
                      onChange={(e) => handleChange('site_keywords', e.target.value)}
                      placeholder="多个关键词用英文逗号分隔"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">OG 图片 URL</label>
                    <input
                      type="url"
                      value={settings['site_og_image'] || ''}
                      onChange={(e) => handleChange('site_og_image', e.target.value)}
                      placeholder="https://example.com/og-image.png"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                </>
              )}

              {activeTab === 'smtp' && (
                <>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">SMTP 服务器</label>
                    <input
                      type="text"
                      value={settings['smtp.host'] || ''}
                      onChange={(e) => handleChange('smtp.host', e.target.value)}
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">SMTP 端口</label>
                    <input
                      type="text"
                      value={settings['smtp.port'] || ''}
                      onChange={(e) => handleChange('smtp.port', e.target.value)}
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">加密方式</label>
                    <CustomSelect
                      value={settings['smtp.encryption'] || 'tls'}
                      onChange={(v) => handleChange('smtp.encryption', v)}
                      options={smtpEncryptionOptions}
                      className="w-full"
                      ariaLabel="SMTP 加密方式"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">发件人邮箱</label>
                    <input
                      type="email"
                      value={settings['smtp.username'] || ''}
                      onChange={(e) => handleChange('smtp.username', e.target.value)}
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">发件人显示名</label>
                    <input
                      type="text"
                      value={settings['smtp.fromName'] || ''}
                      onChange={(e) => handleChange('smtp.fromName', e.target.value)}
                      placeholder="默认同发件人邮箱"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">密码 / 授权码</label>
                    <input
                      type="password"
                      value={
                        settings['smtp.password'] && settings['smtp.password'] !== '***'
                          ? settings['smtp.password']
                          : ''
                      }
                      onChange={(e) => handleChange('smtp.password', e.target.value)}
                      placeholder="留空表示不修改"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>

                  <div className="pt-4 border-t border-border mt-6">
                    <h4 className="text-sm font-medium text-foreground mb-3">发送测试邮件</h4>
                    <div className="flex gap-3">
                      <input
                        type="email"
                        value={testEmail}
                        onChange={(e) => setTestEmail(e.target.value)}
                        placeholder="接收测试邮件的邮箱地址"
                        className="flex-1 px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                      />
                      <button
                        onClick={handleTestSmtp}
                        className="px-4 py-2 bg-muted text-foreground rounded hover:bg-accent transition-colors flex items-center gap-2 text-sm font-medium"
                      >
                        <Mail className="w-4 h-4" />
                        测试发送
                      </button>
                    </div>
                  </div>
                </>
              )}

              {activeTab === 'hexo' && (
                <div className="space-y-6">
                  <div className="bg-primary/10 border border-primary/20 rounded-md p-4">
                    <h4 className="text-sm font-medium text-primary mb-2">Hexo 同步状态</h4>
                    <div className="text-sm text-foreground space-y-1">
                      <p>上次同步时间: {syncStatus?.last_sync_time ? new Date(syncStatus.last_sync_time).toLocaleString() : '从未同步'}</p>
                      <p>数据库发布文章数: {syncStatus?.total_posts ?? 0}</p>
                      <p>本地 Markdown 文件数: {syncStatus?.local_files ?? 0}</p>
                      {syncStatus?.pending_sync && (
                        <p className="text-orange-600 font-medium mt-2">提示：有待同步的变更</p>
                      )}
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">Hexo 根目录（hexo_dir）</label>
                    <input
                      type="text"
                      readOnly
                      value={settings['hexo.hexo_dir'] || ''}
                      className="w-full px-3 py-2 border border-border rounded bg-muted/50 text-muted-foreground text-sm cursor-not-allowed"
                    />
                    <p className="text-xs text-muted-foreground mt-1">由服务端配置文件指定，修改后需重启服务。文章目录默认为 hexo 根目录下的 source/_posts。</p>
                  </div>

                  <div className="flex items-center gap-3">
                    <input
                      id="hexo-auto-sync"
                      type="checkbox"
                      checked={settings['hexo.auto_sync'] === 'true'}
                      onChange={(e) => handleChange('hexo.auto_sync', e.target.checked ? 'true' : 'false')}
                      className="rounded border-border"
                    />
                    <label htmlFor="hexo-auto-sync" className="text-sm text-foreground">
                      保存/更新已发布文章后自动同步到 Hexo
                    </label>
                  </div>

                  <div className="flex items-center gap-3">
                    <input
                      id="hexo-rebuild-after"
                      type="checkbox"
                      checked={settings['hexo.rebuild_after_auto_sync'] === 'true'}
                      onChange={(e) => handleChange('hexo.rebuild_after_auto_sync', e.target.checked ? 'true' : 'false')}
                      className="rounded border-border"
                    />
                    <label htmlFor="hexo-rebuild-after" className="text-sm text-foreground">
                      单篇自动同步后执行 hexo clean + generate
                    </label>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">clean_args（JSON 数组）</label>
                    <textarea
                      value={settings['hexo.clean_args'] ?? ''}
                      onChange={(e) => handleChange('hexo.clean_args', e.target.value)}
                      rows={2}
                      placeholder='例如 ["hexo","clean"] 或 ["pnpm","exec","hexo","clean"]，留空表示不执行'
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground font-mono text-sm"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">generate_args（JSON 数组）</label>
                    <textarea
                      value={settings['hexo.generate_args'] ?? ''}
                      onChange={(e) => handleChange('hexo.generate_args', e.target.value)}
                      rows={2}
                      placeholder='例如 ["hexo","generate"]，留空表示不执行'
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground font-mono text-sm"
                    />
                  </div>

                  <div className="flex flex-wrap gap-3 pt-2">
                    <button
                      type="button"
                      onClick={handleSave}
                      disabled={saving}
                      className="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 transition-colors flex items-center gap-2 text-sm font-medium disabled:opacity-50"
                    >
                      <Save className="w-4 h-4" />
                      {saving ? '保存中...' : '保存 Hexo 设置'}
                    </button>
                    <button
                      type="button"
                      onClick={handleSyncHexo}
                      className="px-4 py-2 bg-muted text-foreground rounded hover:bg-accent transition-colors flex items-center gap-2 text-sm font-medium"
                    >
                      <RefreshCw className="w-4 h-4" />
                      手动触发全量同步
                    </button>
                  </div>
                </div>
              )}

              {activeTab !== 'hexo' && (
                <div className="pt-6">
                  <button
                    onClick={handleSave}
                    disabled={saving}
                    className="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 transition-colors flex items-center gap-2 text-sm font-medium disabled:opacity-50"
                  >
                    <Save className="w-4 h-4" />
                    {saving ? '保存中...' : '保存设置'}
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
