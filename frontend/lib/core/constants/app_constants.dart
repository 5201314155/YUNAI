/// YUNAI 应用常量定义
class AppConstants {
  // 应用信息
  static const String appName = 'YUNAI';
  static const String appDescription = 'AI 社交与创作平台';
  static const String appVersion = '1.0.0';
  
  // API 配置
  static const String baseUrl = 'https://api.yunai.com';
  static const String wsUrl = 'wss://ws.yunai.com';
  static const String mediaUrl = 'https://media.yunai.com';
  
  // 本地存储键名
  static const String accessTokenKey = 'access_token';
  static const String refreshTokenKey = 'refresh_token';
  static const String userInfoKey = 'user_info';
  static const String settingsKey = 'app_settings';
  static const String cacheKey = 'app_cache';
  
  // 网络配置
  static const int connectTimeout = 30000; // 30秒
  static const int receiveTimeout = 30000; // 30秒
  static const int sendTimeout = 30000; // 30秒
  
  // WebSocket 配置
  static const int wsReconnectInterval = 5000; // 5秒
  static const int wsMaxReconnectAttempts = 5;
  static const int wsPingInterval = 30000; // 30秒
  
  // 媒体配置
  static const int maxImageSize = 10 * 1024 * 1024; // 10MB
  static const int maxVideoSize = 100 * 1024 * 1024; // 100MB
  static const int maxAudioSize = 50 * 1024 * 1024; // 50MB
  
  // 聊天配置
  static const int maxMessageLength = 2000;
  static const int maxGroupMembers = 20;
  static const int messagePageSize = 50;
  
  // 角色配置
  static const int maxCharacterName = 20;
  static const int maxCharacterDescription = 500;
  static const int maxCharacterPersonality = 1000;
  
  // 钱包配置
  static const String currencyName = '金币';
  static const int defaultExchangeRate = 10; // 1元 = 10金币
  static const String cardPrefix = '668899'; // YUNAI卡密前缀（6位数字）
  static const int cardLength = 16; // 卡密总长度（16位纯数字）
  
  // 功能开关键名
  static const String enableTxt2img = 'enable_txt2img';
  static const String enableImg2img = 'enable_img2img';
  static const String enableImg2video = 'enable_img2video';
  static const String enableTxt2video = 'enable_txt2video';
  static const String enableRemoveBg = 'enable_remove_bg';
  static const String enableTalkingHead = 'enable_talking_head';
  static const String enableUnifiedGroupModel = 'enable_unified_group_model';
  static const String enableCharacterAutoImage = 'enable_character_auto_image';
  static const String enableSceneSfx = 'enable_scene_sfx';
  static const String enableAiProactiveCall = 'enable_ai_proactive_call';
  static const String enableMomentsAuto = 'enable_moments_auto';
  static const String enableInviteCards = 'enable_invite_cards';
  
  // 用户类型
  static const String userTypeBasic = 'basic';
  static const String userTypeVip = 'vip';
  static const String userTypeCreator = 'creator';
  static const String userTypeAdmin = 'admin';
  
  // 错误码
  static const int successCode = 0;
  static const int invalidRequestCode = 1001;
  static const int unauthorizedCode = 1002;
  static const int forbiddenCode = 1003;
  static const int notFoundCode = 1004;
  static const int internalErrorCode = 5000;
  
  // 动画时长
  static const Duration shortAnimationDuration = Duration(milliseconds: 200);
  static const Duration mediumAnimationDuration = Duration(milliseconds: 300);
  static const Duration longAnimationDuration = Duration(milliseconds: 500);
  
  // 页面配置
  static const double defaultPadding = 16.0;
  static const double smallPadding = 8.0;
  static const double largePadding = 24.0;
  static const double borderRadius = 12.0;
  static const double smallBorderRadius = 8.0;
  static const double largeBorderRadius = 16.0;
  
  // 字体大小
  static const double smallFontSize = 12.0;
  static const double normalFontSize = 14.0;
  static const double mediumFontSize = 16.0;
  static const double largeFontSize = 18.0;
  static const double titleFontSize = 20.0;
  static const double headlineFontSize = 24.0;
  
  // 图标大小
  static const double smallIconSize = 16.0;
  static const double normalIconSize = 24.0;
  static const double mediumIconSize = 32.0;
  static const double largeIconSize = 48.0;
  
  // 头像大小
  static const double smallAvatarSize = 32.0;
  static const double normalAvatarSize = 48.0;
  static const double mediumAvatarSize = 64.0;
  static const double largeAvatarSize = 96.0;
  
  // 正则表达式
  static const String emailRegex = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$';
  static const String phoneRegex = r'^1[3-9]\d{9}$';
  static const String usernameRegex = r'^[a-zA-Z0-9_]{3,20}$';
  static const String passwordRegex = r'^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]{8,}$';
  
  // 支持的文件格式
  static const List<String> supportedImageFormats = ['jpg', 'jpeg', 'png', 'gif', 'webp'];
  static const List<String> supportedVideoFormats = ['mp4', 'mov', 'avi', 'mkv'];
  static const List<String> supportedAudioFormats = ['mp3', 'wav', 'aac', 'm4a'];
  
  // 默认值
  static const String defaultAvatar = 'assets/images/default_avatar.png';
  static const String defaultBackground = 'assets/images/default_background.png';
  static const String defaultCharacterImage = 'assets/images/default_character.png';
}
