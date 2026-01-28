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
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {sources.map((source) => (
              <div
                key={source.source_id}
                className="bg-[#131722] border border-[#2a2e39] rounded-lg p-6 shadow-xl hover:shadow-2xl transition-all hover:border-[#26a69a]/50 group"
              >
                {/* Header */}
                <div className="flex items-start justify-between mb-4">
                  <div className="flex-1">
                    <h3 className="text-lg font-bold text-[#d1d4dc] group-hover:text-[#26a69a] transition-colors mb-1">
                      {source.source_id}
                    </h3>
                    <p className="text-xs text-[#758696]">
                      Created {new Date(source.created_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>

                {/* RSS URL */}
                <div className="mb-4">
                  <label className="text-xs font-medium text-[#758696] mb-1 block">RSS URL</label>
                  <div className="text-sm text-[#d1d4dc] break-all bg-[#1e222d] p-2 rounded border border-[#2a2e39] font-mono">
                    {source.rss_url || '—'}
                  </div>
                </div>

                {/* Selectors */}
                <div className="mb-4">
                  <label className="text-xs font-medium text-[#758696] mb-2 block">CSS Selectors</label>
                  <div className="space-y-2">
                    {source.content_selector && (
                      <div className="text-xs">
                        <span className="text-[#758696]">Content:</span>
                        <span className="ml-2 text-[#d1d4dc] font-mono bg-[#1e222d] px-2 py-1 rounded border border-[#2a2e39]">
                          {source.content_selector}
                        </span>
                      </div>
                    )}
                    {source.author_selector && (
                      <div className="text-xs">
                        <span className="text-[#758696]">Author:</span>
                        <span className="ml-2 text-[#d1d4dc] font-mono bg-[#1e222d] px-2 py-1 rounded border border-[#2a2e39]">
                          {source.author_selector}
                        </span>
                      </div>
                    )}
                    {source.summary_selector && (
                      <div className="text-xs">
                        <span className="text-[#758696]">Summary:</span>
                        <span className="ml-2 text-[#d1d4dc] font-mono bg-[#1e222d] px-2 py-1 rounded border border-[#2a2e39]">
                          {source.summary_selector}
                        </span>
                      </div>
                    )}
                    {source.tags_selector && (
                      <div className="text-xs">
                        <span className="text-[#758696]">Tags:</span>
                        <span className="ml-2 text-[#d1d4dc] font-mono bg-[#1e222d] px-2 py-1 rounded border border-[#2a2e39]">
                          {source.tags_selector}
                        </span>
                      </div>
                    )}
                    {!source.content_selector && !source.author_selector && 
                     !source.summary_selector && !source.tags_selector && (
                      <div className="text-xs text-[#758696] italic">No selectors configured</div>
                    )}
                  </div>
                </div>

                {/* Actions */}
                <div className="flex items-center gap-2 pt-4 border-t border-[#2a2e39]">
                  <button
                    onClick={() => handleAnalyze(source.source_id)}
                    disabled={analyzingId === source.source_id}
                    className="flex-1 px-3 py-2 text-xs font-medium text-white bg-gradient-to-r from-purple-600 to-purple-700 rounded-lg hover:from-purple-700 hover:to-purple-800 focus:outline-none focus:ring-2 focus:ring-purple-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 flex items-center justify-center gap-2"
                    title="Analyze RSS and CSS selectors"
                  >
                    {analyzingId === source.source_id ? (
                      <>
                        <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                        <span>Analyzing...</span>
                      </>
                    ) : (
                      <>
                        <span>🔍</span>
                        <span>Analyze</span>
                      </>
                    )}
                  </button>
                  <button
                    onClick={() => navigate(`/sources/${source.source_id}/edit`)}
                    className="px-3 py-2 text-xs font-medium text-[#d1d4dc] bg-[#1e222d] border border-[#2a2e39] rounded-lg hover:bg-[#252936] hover:border-[#26a69a]/50 focus:outline-none focus:ring-2 focus:ring-[#26a69a] transition-all"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => handleDelete(source.source_id)}
                    disabled={deletingId === source.source_id}
                    className="px-3 py-2 text-xs font-medium text-white bg-gradient-to-r from-[#ef5350] to-[#c62828] rounded-lg hover:from-[#f56565] hover:to-[#d32f2f] focus:outline-none focus:ring-2 focus:ring-[#ef5350] disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                  >
                    {deletingId === source.source_id ? '...' : '🗑️'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </Layout>
  );
}
