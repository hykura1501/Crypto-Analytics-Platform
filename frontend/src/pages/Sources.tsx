import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../api/client';
import type { Source } from '../types';
import Layout from '../components/Layout';

export default function Sources() {
  const navigate = useNavigate();
  const [sources, setSources] = useState<Source[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [analyzingId, setAnalyzingId] = useState<string | null>(null);

  useEffect(() => {
    fetchSources();
  }, []);

  const fetchSources = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getSources();
      setSources(data || []);
    } catch (err) {
      console.error('Error fetching sources:', err);
      const error = err as { response?: { data?: { message?: string } }; message?: string };
      setError(error.response?.data?.message || error.message || 'Failed to load sources');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (sourceId: string) => {
    if (!confirm(`Are you sure you want to delete source "${sourceId}"?`)) {
      return;
    }

    try {
      setDeletingId(sourceId);
      await apiClient.deleteSource(sourceId);
      setSources(sources.filter((s) => s.source_id !== sourceId));
    } catch (err) {
      const error = err as { response?: { data?: { message?: string } }; message?: string };
      alert(error.response?.data?.message || error.message || 'Failed to delete source');
    } finally {
      setDeletingId(null);
    }
  };

  const handleAnalyze = async (sourceId: string) => {
    try {
      setAnalyzingId(sourceId);
      await apiClient.analyzeSource(sourceId);
      alert('Analysis triggered successfully! RSS and CSS selector analysis will be processed in the background.');
      // Refresh sources to get updated selectors
      setTimeout(() => {
        fetchSources();
      }, 3000);
    } catch (err) {
      const error = err as { response?: { data?: { message?: string } }; message?: string };
      alert(error.response?.data?.message || error.message || 'Failed to trigger analysis');
    } finally {
      setAnalyzingId(null);
    }
  };

  return (
    <Layout
      extraActions={
        <button
          onClick={() => navigate('/sources/new')}
          className="px-4 py-2 text-sm font-medium text-white bg-gradient-to-r from-[#26a69a] to-[#1e7a6e] rounded-lg hover:from-[#2db8a8] hover:to-[#268a7a] focus:outline-none focus:ring-2 focus:ring-[#26a69a] transition-all duration-200 shadow-lg"
        >
          + Add Source
        </button>
      }
    >
      {/* Main Content */}
      <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {loading && (
          <div className="flex items-center justify-center py-20">
            <div className="flex flex-col items-center gap-3">
              <div className="w-8 h-8 border-4 border-[#26a69a] border-t-transparent rounded-full animate-spin"></div>
              <div className="text-[#d1d4dc] text-sm font-medium">Loading sources...</div>
            </div>
          </div>
        )}

        {error && (
          <div className="mb-6 bg-[#ef5350] bg-opacity-90 backdrop-blur-sm text-white p-4 rounded-lg border border-[#ef5350] flex items-center justify-between">
            <span>{error}</span>
            <button
              onClick={fetchSources}
              className="ml-4 px-3 py-1 text-sm bg-white bg-opacity-20 hover:bg-opacity-30 rounded transition-all"
            >
              Retry
            </button>
          </div>
        )}

        {!loading && !error && sources.length === 0 && (
          <div className="text-center py-20">
            <div className="inline-block p-4 bg-[#1e222d] rounded-full mb-4">
              <span className="text-4xl">📡</span>
            </div>
            <p className="text-[#758696] mb-4 text-lg">No sources found</p>
            <button
              onClick={() => navigate('/sources/new')}
              className="px-6 py-3 text-sm font-medium text-white bg-gradient-to-r from-[#26a69a] to-[#1e7a6e] rounded-lg hover:from-[#2db8a8] hover:to-[#268a7a] transition-all shadow-lg"
            >
              Create your first source
            </button>
          </div>
        )}

        {!loading && !error && sources && sources.length > 0 && (
          <div className="bg-[#131722] border border-[#2a2e39] rounded-lg shadow-xl overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-[#2a2e39] bg-[#1e222d]">
                    <th className="px-6 py-4 text-xs font-semibold uppercase tracking-wider text-[#758696]">Source ID</th>
                    <th className="px-6 py-4 text-xs font-semibold uppercase tracking-wider text-[#758696]">RSS URL</th>
                    <th className="px-6 py-4 text-xs font-semibold uppercase tracking-wider text-[#758696]">Created At</th>
                    <th className="px-6 py-4 text-xs font-semibold uppercase tracking-wider text-[#758696]">Selectors</th>
                    <th className="px-6 py-4 text-xs font-semibold uppercase tracking-wider text-[#758696] text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#2a2e39]">
                  {sources.map((source) => (
                    <tr key={source.source_id} className="hover:bg-[#1e222d] transition-colors">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm font-medium text-[#d1d4dc]">{source.source_id}</span>
                      </td>
                      <td className="px-6 py-4">
                        <div className="text-sm text-[#d1d4dc] max-w-xs truncate font-mono" title={source.rss_url}>
                          {source.rss_url || '—'}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm text-[#758696]">{new Date(source.created_at).toLocaleDateString()}</span>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex flex-wrap gap-2">
                          {source.content_selector && (
                            <span className="px-2 py-0.5 text-xs rounded bg-[#26a69a]/10 text-[#26a69a] border border-[#26a69a]/20">Content</span>
                          )}
                          {source.author_selector && (
                            <span className="px-2 py-0.5 text-xs rounded bg-blue-500/10 text-blue-400 border border-blue-500/20">Author</span>
                          )}
                          {source.summary_selector && (
                            <span className="px-2 py-0.5 text-xs rounded bg-purple-500/10 text-purple-400 border border-purple-500/20">Summary</span>
                          )}
                          {!source.content_selector && !source.author_selector && !source.summary_selector && (
                            <span className="text-xs text-[#758696] italic">None</span>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 text-right whitespace-nowrap">
                        <div className="flex items-center justify-end gap-2">
                          <button
                            onClick={() => handleAnalyze(source.source_id)}
                            disabled={analyzingId === source.source_id}
                            className="p-2 text-purple-400 hover:bg-purple-500/10 rounded-lg transition-colors disabled:opacity-50"
                            title="Analyze"
                          >
                            {analyzingId === source.source_id ? (
                              <div className="w-4 h-4 border-2 border-purple-400 border-t-transparent rounded-full animate-spin"></div>
                            ) : (
                              <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
                              </svg>
                            )}
                          </button>
                          
                          <button
                            onClick={() => navigate(`/sources/${source.source_id}/edit`)}
                            className="p-2 text-[#26a69a] hover:bg-[#26a69a]/10 rounded-lg transition-colors"
                            title="Edit"
                          >
                            <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            </svg>
                          </button>

                          <button
                            onClick={() => handleDelete(source.source_id)}
                            disabled={deletingId === source.source_id}
                            className="p-2 text-[#ef5350] hover:bg-[#ef5350]/10 rounded-lg transition-colors disabled:opacity-50"
                            title="Delete"
                          >
                            {deletingId === source.source_id ? (
                              <div className="w-4 h-4 border-2 border-[#ef5350] border-t-transparent rounded-full animate-spin"></div>
                            ) : (
                              <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                              </svg>
                            )}
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </Layout>
  );
}
