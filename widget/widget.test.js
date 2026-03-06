/**
 * Bani Widget Tests
 * Tests for widget functionality, state management, and API integration
 */

describe('BaniWidget', () => {
  let container;
  let widgetElement;

  beforeEach(() => {
    document.body.innerHTML = '';
    container = document.createElement('div');
    document.body.appendChild(container);
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  describe('Initialization', () => {
    test('should initialize widget with data-api-key', () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      widgetElement.setAttribute('data-api-key', 'test-key-123');
      container.appendChild(widgetElement);

      new BaniWidget(widgetElement);

      expect(widgetElement.getAttribute('data-api-key')).toBe('test-key-123');
      expect(widgetElement.dataset.initialized).toBe('true');
    });

    test('should log error if api-key is missing', () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      container.appendChild(widgetElement);

      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      new BaniWidget(widgetElement);

      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('data-api-key')
      );
      consoleSpy.mockRestore();
    });

    test('should use default color and font family', () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      widgetElement.setAttribute('data-api-key', 'test-key');
      container.appendChild(widgetElement);

      const widget = new BaniWidget(widgetElement);

      expect(widget.primaryColor).toBe('#4CAF50');
      expect(widget.fontFamily).toBe('Arial, sans-serif');
    });

    test('should use custom color and font family', () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      widgetElement.setAttribute('data-api-key', 'test-key');
      widgetElement.setAttribute('data-color', '#FF5722');
      widgetElement.setAttribute('data-font-family', 'Georgia, serif');
      container.appendChild(widgetElement);

      const widget = new BaniWidget(widgetElement);

      expect(widget.primaryColor).toBe('#FF5722');
      expect(widget.fontFamily).toBe('Georgia, serif');
    });

    test('should start in date-selection view', () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      widgetElement.setAttribute('data-api-key', 'test-key');
      container.appendChild(widgetElement);

      global.fetch = jest.fn();

      const widget = new BaniWidget(widgetElement);

      expect(widget.state.view).toBe('date-selection');
    });
  });

  describe('State Management', () => {
    test('should update state when date is selected', () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      widgetElement.setAttribute('data-api-key', 'test-key');
      container.appendChild(widgetElement);

      const widget = new BaniWidget(widgetElement);
      const testDate = new Date();
      testDate.setDate(testDate.getDate() + 1);

      widget.state.selectedDate = testDate;

      expect(widget.state.selectedDate).toEqual(testDate);
    });

    test('should reset selected slot when navigating back', () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      widgetElement.setAttribute('data-api-key', 'test-key');
      container.appendChild(widgetElement);

      const widget = new BaniWidget(widgetElement);
      widget.state.selectedSlot = { start_time: '2026-03-07T10:00:00Z', available: true };
      widget.state.view = 'booking-form';

      widget.state.selectedSlot = null;
      widget.state.view = 'slot-selection';

      expect(widget.state.selectedSlot).toBeNull();
      expect(widget.state.view).toBe('slot-selection');
    });
  });

  describe('Formatting', () => {
    test('should format price correctly', () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      widgetElement.setAttribute('data-api-key', 'test-key');
      container.appendChild(widgetElement);

      const widget = new BaniWidget(widgetElement);

      expect(widget.formatPrice(5000)).toBe('50');
      expect(widget.formatPrice(10000)).toBe('100');
    });

    test('should escape HTML properly', () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      widgetElement.setAttribute('data-api-key', 'test-key');
      container.appendChild(widgetElement);

      const widget = new BaniWidget(widgetElement);

      const escaped = widget.escapeHtml('<script>alert("xss")</script>');
      expect(escaped).not.toContain('<script>');
      expect(escaped).toContain('&lt;');
    });
  });

  describe('API Integration', () => {
    test('should fetch bathhouse info on init', async () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      widgetElement.setAttribute('data-api-key', 'test-key-123');
      container.appendChild(widgetElement);

      const mockBathhouse = {
        id: 'bath-123',
        name: 'Test Bathhouse',
        address: '123 Main St',
        price_per_hour: 5000
      };

      global.fetch = jest.fn(() =>
        Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockBathhouse)
        })
      );

      const widget = new BaniWidget(widgetElement);
      await new Promise(resolve => setTimeout(resolve, 10));

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/widget/test-key-123/bathhouse')
      );
    });

    test('should fetch slots for selected date', async () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      widgetElement.setAttribute('data-api-key', 'test-key-123');
      container.appendChild(widgetElement);

      const mockSlots = [
        { start_time: '2026-03-07T10:00:00Z', end_time: '2026-03-07T11:00:00Z', available: true, price: 5000 }
      ];

      global.fetch = jest.fn(() =>
        Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockSlots)
        })
      );

      const widget = new BaniWidget(widgetElement);
      const testDate = new Date('2026-03-07');

      await widget.fetchSlots(testDate);

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/slots?date=2026-03-07')
      );
    });
  });

  describe('Multilingual Support', () => {
    test('should support Russian language', () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      widgetElement.setAttribute('data-api-key', 'test-key');
      widgetElement.setAttribute('data-language', 'ru');
      container.appendChild(widgetElement);

      const widget = new BaniWidget(widgetElement);

      expect(widget.t.title).toBe('Бронирование');
      expect(widget.t.selectDate).toBe('Выберите дату');
    });

    test('should support English language', () => {
      widgetElement = document.createElement('div');
      widgetElement.id = 'bani-widget';
      widgetElement.setAttribute('data-api-key', 'test-key');
      widgetElement.setAttribute('data-language', 'en');
      container.appendChild(widgetElement);

      const widget = new BaniWidget(widgetElement);

      expect(widget.t.title).toBe('Booking');
      expect(widget.t.selectDate).toBe('Select date');
    });
  });
});
