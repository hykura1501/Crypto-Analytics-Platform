import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { apiClient } from '../api/client';
import type { CreateSourceRequest, UpdateSourceRequest, Source } from '../types';

export default function SourceForm() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = Boolean(id);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [source, setSource] = useState<Source | null>(null);

  const [formData, setFormData] = useState({
    source_id: '',
    rss_url: '',
    title_tag: '',
    link_tag: '',
    pub_date_tag: '',
    summary_selector: '',
    content_selector: '',
    author_selector: '',
    tags_selector: '',
  });

  useEffect(() => {
    if (isEdit && id) {
      fetchSource(id);
    }
  }, [isEdit, id]);

  const fetchSource = async (sourceId: string) => {
    try {
      setLoading(true);
      const data = await apiClient.getSource(sourceId);
      setSource(data);
      setFormData({
        source_id: data.source_id,
        rss_url: data.rss_url,
        title_tag: data.title_tag || '',
        link_tag: data.link_tag || '',
        pub_date_tag: data.pub_date_tag || '',
        summary_selector: data.summary_selector || '',
        content_selector: data.content_selector || '',
        author_selector: data.author_selector || '',
        tags_selector: data.tags_selector || '',
      });
    } catch (err) {
      const error = err as { response?: { data?: { message?: string } }; message?: string };
      setError(error.response?.data?.message || error.message || 'Failed to load source');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      if (isEdit && id) {
        const updateData: UpdateSourceRequest = {
          rss_url: formData.rss_url || undefined,
          title_tag: formData.title_tag || undefined,
          link_tag: formData.link_tag || undefined,
          pub_date_tag: formData.pub_date_tag || undefined,
          summary_selector: formData.summary_selector || undefined,
          content_selector: formData.content_selector || undefined,
          author_selector: formData.author_selector || undefined,
          tags_selector: formData.tags_selector || undefined,
        };
        await apiClient.updateSource(id, updateData);
      } else {
        const createData: CreateSourceRequest = {
          source_id: formData.source_id,
          rss_url: formData.rss_url,
        };
        await apiClient.createSource(createData);
      }
      navigate('/sources');
    } catch (err) {
      const error = err as { response?: { data?: { message?: string } }; message?: string };
      setError(error.response?.data?.message || error.message || 'Failed to save source');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  if (loading && isEdit && !source) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-500">Loading source...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate('/sources')}
              className="text-gray-600 hover:text-gray-900"
            >
              ← Back to Sources
            </button>
            <h1 className="text-2xl font-bold text-gray-900">
              {isEdit ? 'Edit Source' : 'Create New Source'}
            </h1>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-white shadow rounded-lg">
          <form onSubmit={handleSubmit} className="p-6 space-y-6">
            {error && (
              <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-md">
                {error}
              </div>
            )}

            {/* Basic Information */}
            <div className="border-b border-gray-200 pb-6">
              <h2 className="text-lg font-medium text-gray-900 mb-4">Basic Information</h2>
              <div className="grid grid-cols-1 gap-6">
                <div>
                  <label htmlFor="source_id" className="block text-sm font-medium text-gray-700 mb-2">
                    Source ID <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="text"
                    id="source_id"
                    name="source_id"
                    required={!isEdit}
                    disabled={isEdit}
                    value={formData.source_id}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
                    placeholder="e.g., CoinDesk"
                  />
                  {isEdit && (
                    <p className="mt-1 text-xs text-gray-500">Source ID cannot be changed</p>
                  )}
                </div>

                <div>
                  <label htmlFor="rss_url" className="block text-sm font-medium text-gray-700 mb-2">
                    RSS URL <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="url"
                    id="rss_url"
                    name="rss_url"
                    required
                    value={formData.rss_url}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                    placeholder="https://example.com/rss"
                  />
                </div>
              </div>
            </div>

            {/* RSS Tags */}
            <div className="border-b border-gray-200 pb-6">
              <h2 className="text-lg font-medium text-gray-900 mb-4">RSS Tags</h2>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div>
                  <label htmlFor="title_tag" className="block text-sm font-medium text-gray-700 mb-2">
                    Title Tag
                  </label>
                  <input
                    type="text"
                    id="title_tag"
                    name="title_tag"
                    value={formData.title_tag}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                    placeholder="title (default)"
                  />
                </div>

                <div>
                  <label htmlFor="link_tag" className="block text-sm font-medium text-gray-700 mb-2">
                    Link Tag
                  </label>
                  <input
                    type="text"
                    id="link_tag"
                    name="link_tag"
                    value={formData.link_tag}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                    placeholder="link (default)"
                  />
                </div>

                <div>
                  <label htmlFor="pub_date_tag" className="block text-sm font-medium text-gray-700 mb-2">
                    Pub Date Tag
                  </label>
                  <input
                    type="text"
                    id="pub_date_tag"
                    name="pub_date_tag"
                    value={formData.pub_date_tag}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                    placeholder="pubDate (default)"
                  />
                </div>
              </div>
            </div>

            {/* CSS Selectors */}
            <div>
              <h2 className="text-lg font-medium text-gray-900 mb-4">CSS Selectors</h2>
              <div className="grid grid-cols-1 gap-6">
                <div>
                  <label htmlFor="content_selector" className="block text-sm font-medium text-gray-700 mb-2">
                    Content Selector
                  </label>
                  <input
                    type="text"
                    id="content_selector"
                    name="content_selector"
                    value={formData.content_selector}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 font-mono text-sm"
                    placeholder="article.content"
                  />
                  <p className="mt-1 text-xs text-gray-500">CSS selector for article content</p>
                </div>

                <div>
                  <label htmlFor="author_selector" className="block text-sm font-medium text-gray-700 mb-2">
                    Author Selector
                  </label>
                  <input
                    type="text"
                    id="author_selector"
                    name="author_selector"
                    value={formData.author_selector}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 font-mono text-sm"
                    placeholder=".author"
                  />
                </div>

                <div>
                  <label htmlFor="summary_selector" className="block text-sm font-medium text-gray-700 mb-2">
                    Summary Selector
                  </label>
                  <input
                    type="text"
                    id="summary_selector"
                    name="summary_selector"
                    value={formData.summary_selector}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 font-mono text-sm"
                    placeholder=".summary"
                  />
                </div>

                <div>
                  <label htmlFor="tags_selector" className="block text-sm font-medium text-gray-700 mb-2">
                    Tags Selector
                  </label>
                  <input
                    type="text"
                    id="tags_selector"
                    name="tags_selector"
                    value={formData.tags_selector}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 font-mono text-sm"
                    placeholder=".tags"
                  />
                </div>
              </div>
            </div>

            {/* Actions */}
            <div className="flex items-center justify-end gap-4 pt-6 border-t border-gray-200">
              <button
                type="button"
                onClick={() => navigate('/sources')}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={loading}
                className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? 'Saving...' : isEdit ? 'Update Source' : 'Create Source'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

